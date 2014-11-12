package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"flag"

	"github.com/Sereal/Sereal/Go/sereal"
	"github.com/couchbaselabs/go-slab"
)

import "net/http"
import _ "net/http/pprof"

const (
	GOMAXPROCS                = 20
	ARENA_SMALLEST_SLAB_CLASS = 128
	ARENA_START_SIZE          = 128 * 1024 * 1024
	ARENA_GROW_FACTOR         = 2
	MAX_PACKET_SIZE           = 65536
)

/*****************************
 *        SLAB ARENA         *
 *****************************/
type lockedArena struct {
	arena *slab.Arena
	sync.Mutex
}

func (a *lockedArena) Alloc(size int) []byte {
	a.Lock()
	b := a.arena.Alloc(size)
	a.Unlock()
	return b
}

func (a *lockedArena) AddRef(b []byte) {
	a.Lock()
	a.arena.AddRef(b)
	a.Unlock()
}

func (a *lockedArena) DecRef(b []byte) {
	a.Lock()
	a.arena.DecRef(b)
	a.Unlock()
}

func (a *lockedArena) Stats(m map[string]int64) map[string]int64 {
	a.Lock()
	m2 := a.arena.Stats(m)
	a.Unlock()
	return m2
}

/* global variables */
var Arena lockedArena

/*****************************
 *        MAIN CODE          *
 *****************************/
func main() {
	var proto = flag.String("proto", "udp", "proto TCP or UDP, default UDP")
	var host = flag.String("host", "localhost", "host, default localhost")
	var port = flag.Int("port", 2015, "port, default 2015")

	var netprofile = flag.Bool("netprofile", false, "open socket for remote profiling")
	flag.Parse()

	if *netprofile {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	runtime.GOMAXPROCS(GOMAXPROCS)

	Arena.arena = slab.NewArena(ARENA_SMALLEST_SLAB_CLASS, ARENA_START_SIZE, ARENA_GROW_FACTOR, nil)

	switch *proto {
	case "udp":
		go udpServer(*host + ":" + strconv.Itoa(*port))

	case "tcp":
		go tcpServer(*host + ":" + strconv.Itoa(*port))

	default:
		log.Fatal("unknown protocol")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

func tcpServer(ipport string) {
	listener, err := net.Listen("tcp", ipport)
	if err != nil {
		log.Fatal("UDP listen error:", err)
	}

	defer listener.Close()
	log.Println("TCP server started at '" + ipport + "'")

	/* exiting from udpServer makes dispatcher() exit */
	quit := make(chan struct{})
	done := make(chan struct{})
	dispChan := make(chan []byte)
	//dispChanIn, dispChanOut := MakeBufferedChannel()

	go dispatcher(dispChan, quit, done)
	//go dispatcher(dispChanOut, quit, done)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(conn net.Conn) {
			var err error
			var pkt []byte
			conn_addr := c.RemoteAddr().String()

			defer func() {
				conn.Close()

				if err == io.EOF {
					log.Println("client disconnected", conn_addr)
				} else if err != nil {
					log.Println("failed to receive data from", conn_addr, err)
				} else {
					log.Println("server disconnected client", conn_addr)
				}

				if pkt != nil {
					Arena.DecRef(pkt)
				}
			}()

			log.Println("new connection accepted", conn_addr)

			var ln int
			var size int32
			var offset uint32

			for {
				size = 0
				offset = 0
				pkt = nil

				if err = binary.Read(conn, binary.LittleEndian, &size); err != nil {
					break
				}

				if size < 0 || size > MAX_PACKET_SIZE {
					log.Println("got too long packet length:", size)
					break
				}

				if pkt = Arena.Alloc(int(size)); pkt == nil {
					log.Println("failed to allocate buffer from arena")
					break
				}

				for size > 0 && err == nil {
					ln, err = conn.Read(pkt[offset : offset+uint32(size)])
					size -= int32(ln)
					offset += uint32(ln)
				}

				if err != nil {
					break
				}

				dispChan <- pkt
			}
		}(c)
	}

	log.Println("Shutting down TCP server at '" + ipport + "'")
	close(quit)
	<-done
}

func udpServer(ipport string) {
	conn, e := net.ListenPacket("udp", ipport)
	if e != nil {
		log.Fatal("UDP listen error:", e)
	}

	defer conn.Close()
	log.Println("UDP server started at '" + ipport + "'")

	/* exiting from udpServer makes dispatcher() exit */
	quit := make(chan struct{})
	done := make(chan struct{})
	dispChan := make(chan []byte)

	var ln int
	var err error
	var pkt []byte
	buf := make([]byte, MAX_PACKET_SIZE, MAX_PACKET_SIZE)

	go dispatcher(dispChan, quit, done)

	for {
		pkt = nil

		if ln, _, err = conn.ReadFrom(buf); err != nil {
			log.Println("UDP read error", err)
			continue
		}

		if pkt = Arena.Alloc(ln); pkt == nil {
			log.Println("failed to allocate buffer from arena")
			break
		}

		copy(pkt, buf)
		dispChan <- pkt
	}

	log.Println("Shutting down UDP server at '" + ipport + "'")
	close(quit)
	<-done
}

func dispatcher(dispChanOut <-chan []byte, quit, done chan struct{}) {
	var epochChanIn chan<- []byte
	tick := time.Tick(1 * time.Second)

	/* processEpoch are standalone go routines */

	for {
		select {
		case <-quit:
			if epochChanIn != nil {
				close(epochChanIn)
			}

			done <- struct{}{}
			return

		case epoch := <-tick:
			if epochChanIn != nil {
				close(epochChanIn)
			}

			in, out := MakeBufferedChannel()
			epochChanIn = in

			go processEpoch(epoch, out)

		case pkt := <-dispChanOut:
			epochChanIn <- pkt
		}
	}
}

func processEpoch(epoch time.Time, epochChanOut <-chan []byte) {
	log.Println("start new worker for epoch", epoch.Unix())

	type mergeResult struct {
		count int
		size  int
		err   error
	}

	var recvCount int
	var wg sync.WaitGroup
	var header map[string][]byte

	dec := sereal.NewDecoder()
	mergeChannels := make(map[string]chan<- []byte)

	for {
		pkt, ok := <-epochChanOut
		if !ok {
			break
		}

		recvCount++

		// :15 is to limit search scope
		typeEnded := bytes.IndexByte(pkt[:15], ':')
		versionEnded := bytes.IndexByte(pkt[typeEnded+1:15], ':')
		if versionEnded < 0 {
			log.Println("received bad event")
			Arena.DecRef(pkt)
			continue
		}

		header = nil
		eventType := pkt[:typeEnded]
		event := pkt[typeEnded+versionEnded+2:]

		if err := dec.UnmarshalHeader(event, &header); err != nil {
			log.Println("Failed to read header from event", string(eventType), "err:", err)
			Arena.DecRef(pkt)
			continue
		}

		var eventTypePersona string
		if bytes.Equal(eventType, []byte("WEB")) {
			eventTypePersona = "WEB-" + string(header["__persona__"])
		} else {
			eventTypePersona = string(eventType)
		}

		mergerChanIn, ok := mergeChannels[eventTypePersona]

		if !ok {
			in, out := MakeBufferedChannel()
			mergeChannels[eventTypePersona] = in
			mergerChanIn = in

			go func(mergerChanOut <-chan []byte, eventType string) {
				wg.Add(1)
				defer wg.Done()

				//log.Println("starting new merger goroutine for", eventType)

				merger := sereal.NewMerger()
				merger.DedupeStrings = true
				merger.Compression = sereal.SnappyCompressor{Incremental: true}

				for {
					event, ok := <-mergerChanOut
					if !ok {
						break
					}

					offset := int8(event[0])
					if _, err := merger.Append(event[offset:]); err != nil {
						log.Println("Failed to merge event", eventType, err)
					}

					Arena.DecRef(event)
				}
			}(out, eventTypePersona)
		}

		// it's a trick to not send offset as separate value
		// The ideas is that Arena.DecRef() should be fed by pkt
		// whereas merger.Append() needs pkt[:versionEnded+1]
		pkt[0] = byte(typeEnded + versionEnded + 2)
		mergerChanIn <- pkt
	}

	// first close all channels to let worker start finishing jobs
	for _, ch := range mergeChannels {
		if ch != nil {
			close(ch)
		}
	}

	// wait all workers
	wg.Wait()

	if recvCount > 0 {
		abs_latency := time.Since(epoch)
		latency := time.Since(epoch.Add(1 * time.Second))
		log.Printf("finish processing epoch %d, latency %.2fs, absolute latency %.2f, merged %d\n", epoch.Unix(), latency.Seconds(), abs_latency.Seconds(), recvCount)
	}
}
