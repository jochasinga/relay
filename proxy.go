package robin

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

type HTTPTestServer interface {
	Close()
	CloseClientConnections()
	Start()
	StartTLS()
}

type Proxy struct {
	*httptest.Server
	Latency time.Duration
	Backend HTTPTestServer
}

func (p *Proxy) SetPort(port string) {
	l, err := net.Listen("tcp", "127.0.0.1:" + port)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:" + port); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	p.Server.Listener = l
}

func NewUnstartedProxy(latency time.Duration, backend HTTPTestServer) *Proxy {

	middleFunc := func(w http.ResponseWriter, r *http.Request) {
		<-time.After(latency)
		func(w http.ResponseWriter, r *http.Request) {
			<-time.After(latency)
			s, ok := backend.(*httptest.Server)
			if ok {
				s.Config.Handler.ServeHTTP(w, r)
			} else {
				p := backend.(*Proxy)
				p.Config.Handler.ServeHTTP(w, r)
			}
		}(w, r)
	}
	
	middleServer := httptest.NewUnstartedServer(http.HandlerFunc(middleFunc))

	proxy := &Proxy{
		Server:  middleServer,
		Latency: latency,
		Backend: backend,
	}
	return proxy
}

func NewProxy(latency time.Duration, backend HTTPTestServer) *Proxy {
	proxy := NewUnstartedProxy(latency, backend)
	proxy.Start()
	return proxy
}

