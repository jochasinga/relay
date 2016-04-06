// Copyright 2016 Jo Chasinga. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Overview
--------
Package relay is a standard `httptest.Server` on steroid for end-to-end HTTP testing.
It implements the test server with a delay middleware to simulate latency before
the target test server's handler.

Relay consists of two components, `Proxy` and `Switcher`. They are HTTP Servers
(`httptest.Server`) listening on a system-chosen port on the local loopback 
interface, for use in end-to-end HTTP tests.                                             

Proxy
-----
`Proxy` is used to place before of any `HTTPTestServer` (`httptest.Server`, 
`Proxy`, or `Switcher`) to simulate a proxy server or a connection with some 
network, I/O or CPU latency. It takes a latency unit in `time.Duration` and 
a backend `HTTPTestServer` as arguments.

Switcher
--------
`Switcher` is a basic round-robin-style proxy which takes a latency unit
in `time.Duration` and a `[]HTTPTestServer` to which it will
circulate requests.

Examples
--------
Let's begin setting up a basic `httptest.Server` to send request to.

        var handler = func(w http.ResponseWriter, r *http.Request) {
                fmt.Fprint(w, "Hello world!")
        }

        func TestGet(t *testing.T) {
                ts := httptest.NewServer(http.HandlerFunc(handler))
                resp, _ := http.Get(ts.URL)
                b, _ := ioutil.ReadAll(resp.Body)
                if string(b) != "Hello world!" {
                        t.Error("Response is not as expected.")
                }
        }

Now, let's use Proxy to simulate a slow connection through which a HTTP request
can be sent to the test server.

        // Connection takes 4s to and from the test server
        delay := time.Duration(2) * time.Second
        // Client takes 3s before timing out.
        timeout = time.Duration(3) * time.Second
        // Create a new proxy with a delay and test server backend
        conn := relay.NewProxy(delay, ts)
        client := &Client{ Timeout: timeout }
        start := time.Now()
        _, _ = client.Get(conn.URL)
        elapsed := time.Since(start)
        deviation := time.Duration(50) * time.Millisecond

        if elapsed >= timeout + deviation {
                t.Error("Client took too long to time out.")
        }

Note that the latency will double because of the round trip to and from the server.

`Proxy` can be placed in front of another proxy, and vice versa. So you can create a 
chain of test proxies this way:

        delay := time.Duration(1) * time.Second
        ts := httptest.NewServer(http.HandlerFunc(handler))
        p3 := relay.NewProxy(delay, ts)
        p2 := relay.NewProxy(delay, p3)
        p1 := relay.NewProxy(delay, p2)
        start := time.Now()
        _, _ := http.Get(p1.URL)
        elapsed := time.Since(start)
        if elapsed >= time.Duration(6) * time.Second + time.Duration(10) * time.Millisecond {
                t.Error("Client took longer than expected.")
        }

Each hop to and from the target server will be delayed for one second.

`Switcher` works similarly to a proxy, except it "switches" between several backend servers
for each request in a round-robin fashion.

        ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       fmt.Fprint(w, "Hello world!")
        }))
        ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       fmt.Fprint(w, "Hello mars!")
        }))
        ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       fmt.Fprint(w, "Hello pluto!")
        }))
        p := relay.NewProxy(delay, ts3)
        sw := relay.NewSwitcher(delay, []HTTPTestServer{ts1, ts2, p})

        resp1, _ := http.Get(sw.URL) // hits ts1
        resp2, _ := http.Get(sw.URL) // hits ts2
        resp3, _ := http.Get(sw.URL) // hits p, which eventually hits ts3
        resp4, _ := http.Get(sw.URL) // hits ts1 again
*/
package relay

