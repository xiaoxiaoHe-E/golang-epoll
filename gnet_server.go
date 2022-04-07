package main

import (
	"bytes"
	"log"

	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
)

type echoServer struct {
	*gnet.EventServer
	pool *goroutine.Pool
}

func (es *echoServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	// data := []byte("Anybody can become angry - that is easy, but to be angry with the right person and to the right degree and at the right time and for the right purpose, and in the right way - that is not within everybody's power and is not easy.")
	// var resp_buffer bytes.Buffer
	// for j := 0; j < 4096; j++ { //about 1M
	// 	resp_buffer.Write(data)
	// }
	// out = resp_buffer.Bytes()
	//fmt.Printf("len out: %d\n", len(out))
	_ = es.pool.Submit(func() {

		for {
			//time.Sleep(5 * time.Microsecond)
			data := []byte("Anybody can become angry - that is easy, but to be angry with the right person and to the right degree and at the right time and for the right purpose, and in the right way - that is not within everybody's power and is not easy.")
			var resp_buffer bytes.Buffer
			for j := 0; j < 4096; j++ { //about 1M
				resp_buffer.Write(data)
			}

			c.AsyncWrite(resp_buffer.Bytes())

		}

	})
	return
}

func main() {
	p := goroutine.Default()
	defer p.Release()
	echo := &echoServer{pool: p}
	log.Fatal(gnet.Serve(echo, "tcp://:9000", gnet.WithMulticore(true)))
}
