package robin

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

type Switcher struct {
	*httptest.Server
	Latency time.Duration
//	Last HTTPTestServer
//	Current HTTPTestServer
	Backends []HTTPTestServer
}

func (s *Switcher) SetPort(port string) {
	l, err := net.Listen("tcp", "127.0.0.1:" + port)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:" + port); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	s.Listener = l
}

var currentServerId = 0

func backendGenerator(backends []HTTPTestServer) <-chan HTTPTestServer {
	c := make(chan HTTPTestServer)
	id := currentServerId
	go func() <-chan HTTPTestServer {
		defer close(c)
		c <- backends[id]
		if currentServerId == len(backends) - 1 {
			currentServerId = 0
			return c
		}
		currentServerId++
		return c
	}()
	return c
}

func NewUnstartedSwitcher(latency time.Duration, backends []HTTPTestServer) *Switcher {

	middleFunc := func(w http.ResponseWriter, r *http.Request) {
		<-time.After(latency)
		func(w http.ResponseWriter, r *http.Request) {
			<-time.After(latency)
			backend := <-backendGenerator(backends)
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

	sw := &Switcher{
		Server:  middleServer,
		Latency: latency,
		Backends: backends,
	}
	return sw
}

func NewSwitcher(latency time.Duration, backends []HTTPTestServer) *Switcher {
	sw := NewUnstartedSwitcher(latency, backends)
	sw.Start()
	return sw
}


