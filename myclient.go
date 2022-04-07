package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	flag.Parse()
	addr := "127.0.0.1:9000"
	run_time := 15
	log.Printf("Connct to %s", addr)

	log.Printf("connect init success")
	i := 0
	st := time.Now()

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		fmt.Println("failed to connect", err)
		return
	}
	hello := "hello " + strconv.Itoa(i) + "\n"
	fmt.Printf("message to send: %s", hello)
	conn.Write([]byte(hello))

	var buf = make([]byte, 933892)
	result := bytes.NewBuffer(nil)
	read_num := 0
	data_size := 933888

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("failed to read from connection %v, read len %d", err, n)
			conn.Close()
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

	fmt.Printf("runtime: %ds, total block: %d\n", run_time, i)
}
