// Copyright 2016 Jo Chasinga. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package relay

import (
	"net/http"
	"net/http/httptest"
	"time"
)

// HTTPTestServer is an interface which all instances including httptest.Server implement.
type HTTPTestServer interface {
	Close()
	CloseClientConnections()
	Start()
	StartTLS()
}

// A Proxy is an HTTPTestServer placed in front of another
// HTTPTestServer to simulate a real proxy server or a connection with latency.
type Proxy struct {
	*httptest.Server
	Latency time.Duration
	Backend HTTPTestServer
}

// Close shuts down the proxy and blocks until all outstanding requests
// on this proxy have completed.
func (p *Proxy) Close() {
	p.Server.Close()
}

// Start starts a proxy from NewUnstartedProxy.
func (p *Proxy) Start() {
	p.Server.Start()
}

// TODO: Write test for SetPort().
// SetPort optionally sets a local port number a Proxy should listen on.
// It should be set on an unstarted proxy only.
/*
func (p *Proxy) setPort(port string) {
	l, err := net.Listen("tcp", "127.0.0.1:" + port)
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:" + port); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	p.Server.Listener = l
}
*/

// NewUnstartedProxy Start an unstarted proxy instance.
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

// NewProxy starts and run a proxy instance.
func NewProxy(latency time.Duration, backend HTTPTestServer) *Proxy {
	proxy := NewUnstartedProxy(latency, backend)
	proxy.Start()
	return proxy
}

