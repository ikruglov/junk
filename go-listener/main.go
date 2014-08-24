package main

import (
	"bytes"
	"log"
	"net"
	"runtime"
	"sync"
	"syscall"
	"time"

	"flag"
	"os"
	"os/signal"
	//"runtime/pprof"

	"github.com/Sereal/Sereal/Go/sereal"
	"github.com/couchbaselabs/go-slab"
)

import "net/http"
import _ "net/http/pprof"

const (
	ARENA_SIZE      = 1024 * 1024 * 1024
	MAX_PACKET_SIZE = 65536
	MAX_CHAN_SIZE   = 500000
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

var Arena lockedArena
var BigArena lockedArena

//var MergerArena lockedArena

/*****************************
 *        MAIN CODE          *
 *****************************/
func main() {
	//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	//var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	var netprofile = flag.Bool("netprofile", false, "open socket for remote profiling")
	flag.Parse()

	if *netprofile {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	//	if *cpuprofile != "" {
	//		f, err := os.Create(*cpuprofile)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		pprof.StartCPUProfile(f)
	//		defer pprof.StopCPUProfile()
	//	}

	runtime.GOMAXPROCS(20)
	Arena.arena = slab.NewArena(MAX_PACKET_SIZE, ARENA_SIZE, 2, nil)
	BigArena.arena = slab.NewArena(1024*1024, ARENA_SIZE, 2, nil)

	proto := "udp"
	ipport := ":2005"

	switch proto {
	case "udp":
		go udpServer(ipport)

	default:
		log.Fatal("unknown protocol")
	}

	// setup signal handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	//	if *memprofile != "" {
	//		f, err := os.Create(*memprofile)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//		pprof.WriteHeapProfile(f)
	//		f.Close()
	//		return
	//	}
}

func udpServer(ipport string) {
	conn, e := net.ListenPacket("udp", ipport)
	if e != nil {
		log.Fatal("UDP listen error:", e)
	}

	log.Println("UDP server started at '" + ipport + "'")

	w := make(chan []byte, MAX_CHAN_SIZE)
	quit := make(chan struct{})
	done := make(chan struct{})
	go dispatcher(w, quit, done)

	//for i := 0; i < 100000; i++ {
	for {
		pkt := Arena.Alloc(MAX_PACKET_SIZE)
		// TODO check pkt == nil

		ln, _, err := conn.ReadFrom(pkt)
		if err != nil {
			panic("UDP read error")
			log.Println("UDP read error", err)
			continue
		}

		w <- pkt[:ln]
	}

	log.Println("Shutting down..")
	close(quit)
	<-done
}

func dispatcher(w chan []byte, quit, done chan struct{}) {
	var ch chan []byte
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-quit:
			log.Println("!!!")
			done <- struct{}{}
			return

		case epoch := <-tick:
			if ch != nil {
				close(ch)
			}

			ch = make(chan []byte, MAX_CHAN_SIZE)
			go processEpoch(epoch, ch)

		case pkt := <-w:
			ch <- pkt
		}
	}
}

func processEpoch(epoch time.Time, ch chan []byte) {
	log.Println("start new worker for epoch", epoch.Unix())

	type mergeResult struct {
		count int
		size  int
		err   error
	}

	var recvCount int
	dec := sereal.NewDecoder()
	done := make(chan mergeResult)
	mergeChannels := make(map[string]chan []byte)

	for {
		pkt, ok := <-ch
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

		eventType := pkt[:typeEnded]
		event := pkt[typeEnded+versionEnded+2:]
		header := make(map[string][]byte)

		if err := dec.UnmarshalHeader(event, &header); err != nil {
			log.Println("Failed to read header from event", string(eventType), "err:", err)
			Arena.DecRef(pkt)
			continue
		}

		var eventTypePersona string
		if bytes.Equal(eventType, []byte("WEB")) {
			eventTypePersona = "WEB-" + string(header["persona"])
		} else {
			eventTypePersona = string(eventType)
		}

		mch, ok := mergeChannels[eventTypePersona]

		if !ok {
			mch = make(chan []byte, MAX_CHAN_SIZE)
			mergeChannels[eventTypePersona] = mch

			go func(ch chan []byte, eventType string, done chan mergeResult) {
				var err error
				var count int

				mbuf := BigArena.Alloc(64 * 1024 * 1024)
				merger := sereal.NewMerger()
				merger.SetBuf(mbuf[:5])
				//merger.ExpectedSize = 128 * 1024 * 1024 // 128 MB

				for {
					event, ok := <-ch
					if !ok {
						break
					}

					offset := int8(event[0])
					if cnt, err := merger.Append(event[offset:]); err != nil {
						log.Println("Failed to merge event", eventType, err)
					} else {
						count += cnt
					}

					Arena.DecRef(event)
				}

				res, err := merger.Finish()
				if err == nil {
					done <- mergeResult{count, len(res), nil}
				} else {
					done <- mergeResult{0, 0, err}
				}

				BigArena.DecRef(mbuf)
			}(mch, eventTypePersona, done)
		}

		// it's a trick to not send offset as separate value
		// The ideas is that Arena.DecRef() should be fed by pkt
		// whereas merger.Append() needs pkt[:versionEnded+1]
		pkt[0] = byte(typeEnded + versionEnded + 2)
		mch <- pkt
		// TODO fix possible memleak here
	}

	// first close all channels to let worker start finishing jobs
	for _, ch := range mergeChannels {
		if ch != nil {
			close(ch)
		}
	}

	// now, wait for all results

	mergedSize := 0
	mergedCount := 0
	for i := 0; i < len(mergeChannels); i++ {
		mresult := <-done
		mergedSize += mresult.size
		mergedCount += mresult.count
		//log.Println(epoch.Unix(), t, mresult.count)
	}

	nextepoch := epoch.Add(1 * time.Second)
	latency := time.Since(nextepoch)
	log.Printf("finish processing epoch %d, latency %s, merged %d, mergedSize: %.2fMB\n", epoch.Unix(), latency, mergedCount, float64(mergedSize)/1024/1024)
	//log.Printf("finish processing epoch %d, received %d\n", epoch.Unix(), recvCount)
}
