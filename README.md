## A Epoll Demo in Golang

gnet_server.go: a simple demo server using gnet asyncwrite

myepoll.go: a demo server using epoll

myclient.go: a client to receive message from server

### Test Result
data block len: 933888

Myepoll:  runtime: 15s, total block: 21567, more stable


gnet:  runtime: 15s, total block: 26236, variance is larger, sometimes up to 40000+, sometime 15000

