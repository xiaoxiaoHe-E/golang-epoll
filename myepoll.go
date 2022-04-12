package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

var epoller *epoll

type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func MkEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

func (e *epoll) Add(conn net.Conn, event_type uint32) error {
	// Extract file descriptor associated with the connection
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: event_type, Fd: int32(fd)})
	if err != nil {
		panic(err)
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	if len(e.connections)%100 == 0 {
		log.Printf("add total number of connections: %v\n", len(e.connections))
	}
	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	if len(e.connections)%100 == 0 {
		log.Printf("total number of connections: %v\n", len(e.connections))
	}
	return nil
}

func (e *epoll) Modify(conn net.Conn, event_type uint32) error {
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_MOD, fd, &unix.EpollEvent{Events: event_type, Fd: int32(fd)})
	if err != nil {
		return err
	}
	return nil
}

func (e *epoll) Wait() ([]net.Conn, []uint32, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		return nil, nil, err
	}
	e.lock.RLock()
	defer e.lock.RUnlock()
	var connections []net.Conn
	var event_type []uint32
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		connections = append(connections, conn)
		event_type = append(event_type, events[i].Events)
	}
	return connections, event_type, nil
}
func socketFD(conn net.Conn) int {
	//Get the socketFD of the conn
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	//if tls {
	//	tcpConn = reflect.Indirect(tcpConn.Elem())
	//}
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func main() {

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}
	go start()
	for {
		conn, e := ln.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				log.Printf("accept temp err: %v", ne)
				continue
			}
			log.Printf("accept err: %v", e)
			return
		}
		fmt.Println("accept successful....")
		if err := epoller.Add(conn, unix.EPOLLIN); err != nil {
			log.Printf("failed to add connection %v", err.Error())
			conn.Close()
		}
	}
}

func start() {
	var buf = make([]byte, 20)
	var resp_cnt int = 0
	for {
		connections, event_type, err := epoller.Wait()
		if err != nil {
			log.Printf("failed to epoll wait %v", err)
			continue
		}

		for i, conn := range connections {
			if conn == nil {
				break
			}

			if event_type[i]&unix.EPOLLOUT != 0 { // an EPOLLOUT event
				fmt.Printf("do write: %d\n", resp_cnt)
				resp_cnt++
				data := []byte("Anybody can become angry - that is easy, but to be angry with the right person and to the right degree and at the right time and for the right purpose, and in the right way - that is not within everybody's power and is not easy.")
				var resp_buffer bytes.Buffer
				for j := 0; j < 4096; j++ { //about 1M 933888
					resp_buffer.Write(data)
				}

				n, err := conn.Write(resp_buffer.Bytes())
				fmt.Printf("write len: %d\n", n)
				if err != nil {
					log.Printf("failed to write %v\n", err)
					conn.Close()
					break
				}
			} else if event_type[i]&unix.EPOLLIN != 0 { // an EPOLLIN event
				if _, err := conn.Read(buf); err != nil {
					if err := epoller.Remove(conn); err != nil {
						log.Printf("failed to remove %v\n", err)
					}
					conn.Close()
					break
				}
				if err := epoller.Modify(conn, unix.EPOLLOUT); err != nil {
					// modify it to an out event
					log.Printf("faild to modify the event %v\n", err)
					conn.Close()
					break
				}
			}

		}
	}
}
