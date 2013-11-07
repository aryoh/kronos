package main

// http://mmcgrana.github.io/2012/09/getting-started-with-go-on-heroku.html
// https://github.com/buger/gor/blob/master/replay/request_stats.go#L47

import (
	"flag"
	"fmt"
	zmq "github.com/alecthomas/gozmq"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	"util/pinba"
)

type Message struct {
	ts       int
	requests []pinba.Request
}

func receiver(in_addr *string, out chan<- Message) {
	context, _ := zmq.NewContext()
	defer context.Close()

	subscriber, _ := context.NewSocket(zmq.SUB)
	subscriber.Connect(*in_addr)
	subscriber.SetSockOptString(zmq.SUBSCRIBE, "")
	defer subscriber.Close()

	fmt.Printf("Waiting for data on %v\n", *in_addr)
	for {
		// Получение исходных данных от pinba2zmq
		msgbytes, err := subscriber.Recv(0)
		if err != nil {
			fmt.Println("Receive Error:", err.Error())
		} else {
			ts, requests := pinba.Decode(&msgbytes)
			fmt.Printf("Receive: %v, %v\n", ts, len(requests))
			out <- Message{ts: ts, requests: requests}
		}
	}
}

func process(message *Message) {
	t := time.Now()
	for _, request := range message.requests {
		if request.Status == nil {
			continue
		}
		if *dump_type == "php" {
			fmt.Printf("Request: %v - %v, %v, %v\n",
				*request.Status, *request.Hostname, *request.ServerName, *request.ScriptName)
			fmt.Printf(" Data: doc_size: %v, mem: %v, cnt: %v\n",
				*request.DocumentSize, *request.MemoryPeak, *request.RequestCount)
			fmt.Printf(" Time: req_time: %v, cpu_utime: %v, cpu_stime: %v\n",
				*request.RequestTime, *request.RuUtime, *request.RuStime)
			fmt.Printf(" Timers:\n")
			for _, timer := range request.Timers() {
				fmt.Printf(" - %v, %v\n", timer.Value, timer.Count)
				for k, v := range timer.Tags {
					fmt.Printf("   * %v: %v\n", k, v)
				}
				fmt.Printf("\n")
			}
		}

		if *dump_type == "nginx" {
			fmt.Printf("HTTP % 4d: [%v] % 7d %3.2fms %v %v\n",
				*request.Status, *request.Hostname, *request.DocumentSize, *request.RequestTime*1000,
				*request.ServerName, *request.ScriptName)
			for _, timer := range request.Timers() {
				fmt.Printf(" - %v, %v\n", timer.Value, timer.Count)
				for k, v := range timer.Tags {
					fmt.Printf("   * %v: %v\n", k, v)
				}
				fmt.Printf("\n")
			}
		}

		//fmt.Printf("\n---\n")
	}
	fmt.Printf("Processed: %v, %v (%v)\n\n", message.ts, len(message.requests), time.Now().Sub(t))
}

var (
	in_addr   = flag.String("in", "", "incoming socket")
	dump_type = flag.String("type", "php", "incoming socket")

	server      = flag.String("server", "", "server")
	hostname    = flag.String("hostname", "", "hostname")
	script_name = flag.String("script_name", "", "script_name")
)

// kronos_dump --in=tcp://172.16.5.130:5003 --type=nginx

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(2)

	input := make(chan Message)
	go receiver(in_addr, input)

	exit := make(chan bool)
	signal_channel := make(chan os.Signal)
	signal.Notify(signal_channel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func(exit_channel chan bool) {
		fmt.Printf("\nSignal caught - %v!\n", <-signal_channel)
		exit_channel <- true
	}(exit)

	exit_signal := false
	for exit_signal == false {
		select {
		case exit_signal = <-exit:
			fmt.Println("W: interrupt received, killing server...")

		case m := <-input:
			if &m != nil {
				process(&m)
			}
		}
	}
}
