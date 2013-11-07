package main

import (
	"bytes"
	"compress/zlib"
	"flag"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"log"
	"os"
	"time"
	//"os/signal"
	"net"
)

func handle(conn *net.UDPConn) {
	var buf [65536]byte
	rlen, _, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		logger.Printf("Error on conn.ReadFrom, %v", err)
		panic(err)
	}
	if rlen == 65536 {
		logger.Printf("Read exactly 65536 bytes!\n")
	}

	buffer.Write(buf[0:rlen])
	buffer.Write([]byte{0xa, 0x2d, 0x2d, 0xa}) // Delimeter
	counter += 1
}

func sender(out_addr *string) {
	context, _ := zmq.NewContext()
	defer context.Close()
	publisher, err := context.NewSocket(zmq.PUB)
	if err != nil {
		logger.Printf("Error on context.NewSocket(zmq.PUB), %v", err)
		panic(err)
	}
	defer publisher.Close()
	publisher.Bind(*out_addr)
	publisher.SetHWM(1)
	publisher.SetSwap(512 * 1024 * 1024)

	c := time.Tick(1 * time.Second)
	for now := range c {
		logger.Printf("%v: %v, %v, %v \n", time.Now(), now.Unix(), len(buffer.Bytes()), counter)
		if counter > 0 {
			var b bytes.Buffer
			w := zlib.NewWriter(&b)
			w.Write(buffer.Bytes()[:len(buffer.Bytes())-4])
			w.Close()
			publisher.Send([]byte(fmt.Sprintf("%d\n%s", now.Unix(), b.Bytes())), 0)
		}
		buffer = *bytes.NewBuffer([]byte{})
		counter = 0
	}
}

var (
	buffer   bytes.Buffer
	logger   log.Logger
	counter  = 0
	in_addr  = flag.String("in", "", "incoming socket")
	out_addr = flag.String("out", "", "outcoming socket")
	log_file = flag.String("log", "", "log filename")
)

func main() {
	flag.Parse()

	fh, err := os.OpenFile(*log_file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Error on os.Create(*log_file), %v", err)
		panic(err)
	}
	defer fh.Close()
	logger = *log.New(fh, "", 0)

	addr, err := net.ResolveUDPAddr("udp4", *in_addr)
	if err != nil {
		fmt.Printf("Error on net.ResolveUDPAddr, %v", err)
		panic(err)
	}
	sock, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Printf("Error on net.ListenUDP, %v", err)
		panic(err)
	}
	logger.Printf("Start listening on %v\n", *in_addr)
	go sender(out_addr)
	for {
		handle(sock)
	}
}
