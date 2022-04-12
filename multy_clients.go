package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
)

const (
	addr     = "127.0.0.1:9000"
	run_time = 10
	seg_num  = 256
)

func main() {
	flag.Parse()

	c := make(chan int, seg_num+1)

	for i := 0; i < seg_num; i++ {
		go run_client(c, i)
	}

	total_read_block := 0
	for i := 0; i < seg_num; i++ {
		re, ok := <-c
		if !ok {
			fmt.Println("failed to read from channel")
			break
		}
		total_read_block += re
	}
	fmt.Printf("seg_num: %d, runtime: %ds, total block: %d\n", seg_num, run_time, total_read_block)
	return
}

func run_client(c chan int, func_num int) {

	i := 0
	st := time.Now()

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		fmt.Println("failed to connect", err)
		return
	}
	hello := "hello " + strconv.Itoa(i) + "\n"
	//fmt.Printf("message to send: %s", hello)
	conn.Write([]byte(hello))

	var buf = make([]byte, 933892)
	result := bytes.NewBuffer(nil)
	read_num := 0
	data_size := 933888

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("failed to read from connection %v, read len %d\n", err, n)
			conn.Close()
			close(c)
			return
		}

		read_num += n
		if read_num < data_size {
			result.Write(buf[0:n])
		} else {
			result.Write(buf[0 : n-(read_num-data_size)])
			fmt.Printf("read block %d, read len: %d\n", i, result.Len())
			i++
			result.Reset()
			result.Write(buf[n-(read_num-data_size) : n])
			read_num = read_num - data_size
		}

		if int(time.Since(st).Seconds()) > run_time {
			break
		}
	}

	fmt.Printf("go: %d, total block: %d\n", func_num, i)
	c <- i
	return
}
