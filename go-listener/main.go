package main

import (
	"bytes"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/Sereal/Sereal/Go/sereal"
)

func main() {
	proto := "udp"
	ipport := ":2005"

	runtime.GOMAXPROCS(4)

	switch proto {
	case "udp":
		udpServer(ipport)

	default:
		log.Fatal("unknown protocol")
	}
}

func udpServer(ipport string) {
	conn, e := net.ListenPacket("udp", ipport)
	if e != nil {
		log.Fatal("UDP listen error:", e)
	}

	log.Println("UDP server started at '" + ipport + "'")

	var buf [65536]byte
	w := make(chan []byte)
	quit := make(chan struct{})
	done := make(chan struct{})
	go worker(w, quit, done)

	for {
		ln, _, err := conn.ReadFrom(buf[:])
		if err != nil {
			log.Println("UDP read error", err)
			continue
		}

		w <- buf[:ln]
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

			ch = make(chan []byte, 4096)
			go processEpoch(epoch, ch)

		case pkt := <-w:
			ch <- pkt
		}
	}
}

func processEpoch(epoch time.Time, ch chan []byte) {
	log.Println("start new worker for epoch", epoch.Unix())

	var msg []byte
	var recvCount int

	type mergerData struct {
		merger *sereal.Merger
		count  int
	}

	mergerMap := make(map[string]*mergerData)

	for {
		msg = <-ch
		if msg == nil {
			break
		}

		recvCount++
		data := bytes.SplitN(msg, []byte(":"), 3)
		if len(data) != 3 {
			log.Println("received bad event")
			continue
		}

		eventType := string(data[0])
		md, ok := mergerMap[eventType]
		if !ok {
			md = &mergerData{sereal.NewMerger(), 0}
			mergerMap[eventType] = md
		}

		cnt := 1
		//cnt, err := md.merger.Append(data[2])
		//if err != nil {
		//    log.Println("failed to merge event", eventType, err)
		//    continue
		//}

		md.count += cnt
	}

	mergedCount := 0
	for t, md := range mergerMap {
		log.Println(t, md.count)
		mergedCount += md.count
	}

	log.Printf("finish processing epoch %d, received %d, merged %d\n", epoch.Unix(), recvCount, mergedCount)
}
