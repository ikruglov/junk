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
	"runtime/pprof"

	"github.com/Sereal/Sereal/Go/sereal"
	"github.com/couchbaselabs/go-slab"
)

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

//var MergerArena lockedArena

/*****************************
 *        MAIN CODE          *
 *****************************/
func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	runtime.GOMAXPROCS(20)
	Arena.arena = slab.NewArena(MAX_PACKET_SIZE, ARENA_SIZE, 2, nil)

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

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
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
	go worker(w, quit, done)

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

func worker(w chan []byte, quit, done chan struct{}) {
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

	type mergerData struct {
		merger *sereal.Merger
		ch     chan []byte
		count  int
	}

	var recvCount int
	done := make(chan bool)
	dec := sereal.NewDecoder()
	mergerMap := make(map[string]*mergerData)

	for {
		pkt, ok := <-ch
		if !ok {
			break
		}

		recvCount++

		data := bytes.SplitN(pkt, []byte(":"), 3)
		if len(data) != 3 {
			panic("received bad event")
			log.Println("received bad event")
			continue
		}

		header := make(map[string][]byte)
		if err := dec.UnmarshalHeader(data[2], &header); err != nil {
			log.Println(err)
			continue
		}

		var eventTypePersona string
		if bytes.Equal(data[0], []byte("WEB")) {
			eventTypePersona = "WEB-" + string(header["persona"])
		} else {
			eventTypePersona = string(data[0])
		}

		switch eventTypePersona {
		case "WEB-app", "WEB-xml", "WEB-sessapp":
			mdata, ok := mergerMap[eventTypePersona]
			if !ok {
				mdata = &mergerData{
					sereal.NewMerger(),
					make(chan []byte, MAX_CHAN_SIZE),
					0,
				}

				mdata.merger.ExpectedSize = 512 * 1024 * 1024
				mergerMap[eventTypePersona] = mdata
				go func(mdata *mergerData, eventType string, done chan bool) {
					for {
						event, ok := <-mdata.ch
						if !ok {
							break
						}

						pos1 := bytes.IndexByte(event, ':')
						pos2 := bytes.IndexByte(event[pos1+1:], ':')

						cnt, err := mdata.merger.Append(event[pos1+pos2+2:])
						if err != nil {
							log.Println("Failed to merge event", eventType, err)
							panic("!!!!! 1")
						}

						mdata.count += cnt
						Arena.DecRef(event)
					}

					done <- true
				}(mdata, eventTypePersona, done)
			}

			mdata.ch <- pkt

		default:
			mdata, ok := mergerMap[eventTypePersona]
			if !ok {
				mdata = &mergerData{sereal.NewMerger(), nil, 0}
				mergerMap[eventTypePersona] = mdata
			}

			cnt, err := mdata.merger.Append(data[2])
			if err != nil {
				log.Println("failed to merge event", eventTypePersona, err)
				panic("!!!!! 2")
			}

			mdata.count += cnt
			Arena.DecRef(pkt)
		}
	}

	mergedCount := 0
	for t, md := range mergerMap {
		if md.ch != nil {
			close(md.ch)
			<-done
		}

		md.merger.Finish()
		mergedCount += md.count
		log.Println(epoch.Unix(), t, md.count)
	}

	nextepoch := epoch.Add(1 * time.Second)
	latency := time.Since(nextepoch)
	log.Printf("finish processing epoch %d, latency %s, received %d, merged %d\n", epoch.Unix(), latency, recvCount, mergedCount)
	//log.Printf("finish processing epoch %d, received %d\n", epoch.Unix(), recvCount)
}
