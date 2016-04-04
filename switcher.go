package relay

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
	"sync"
	"runtime"
)

// A Switcher is an HTTPTestServer placed in front of another HTTPTestServer's
// to simulate a round-robin load-balancer.
type Switcher struct {
	*httptest.Server
	Latency time.Duration
	Backends []HTTPTestServer
}

var (
	currentServerId int
	mutex = &sync.Mutex{}
)

// SetPort optionally sets a local port number a Switcher should listen on.
// It should be set on an unstarted switcher only.
func (s *Switcher) SetPort(port string) {
	l, err := net.Listen("tcp", "127.0.0.1:" + port)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:" + port); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	s.Listener = l
}

// backendGenerator keeps track of the backend servers circulation.
func backendGenerator(backends []HTTPTestServer) <-chan HTTPTestServer {
	c := make(chan HTTPTestServer)
	go func() <-chan HTTPTestServer {
		defer close(c)
		mutex.Lock()
		c <- backends[currentServerId]
		mutex.Unlock()
		if currentServerId == len(backends) - 1 {
			currentServerId = 0
			return c
		}
		mutex.Lock()
		currentServerId++
		mutex.Unlock()
		runtime.Gosched()
		return c
	}()
	return c
}

// NewUnstartedProxy Start an unstarted proxy instance. 
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

// NewProxy starts and run a proxy instance. 
func NewSwitcher(latency time.Duration, backends []HTTPTestServer) *Switcher {
	sw := NewUnstartedSwitcher(latency, backends)
	sw.Start()
	return sw
}


