// Copyright 2016 Jo Chasinga. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package relay is a standard httptest.Server on steroid for end-to-end HTTP testing.

The basic building blocks are a Proxy and Switcher, which are HTTP Servers
(httptest.Server) listening on a system-chosen port on the local loopback 
interface, for use in end-to-end HTTP tests.                                             

A Proxy is used to place in front of any HTTPTestServer (httptest.Server, 
Proxy, or Switcher) to simulate a proxy server or a connection with some 
network, I/O or CPU latency. It takes a latency unit in time.Duration and 
a backend HTTPTestServer as arguments.

A Switcher is a basic round-robin-style proxy which takes a latency unit
in time.Duration and a slice of backend HTTPTestServer's among which it will
circulate requests.

Let's begin setting up a basic httptest.Server to test against.

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

Now, let's simulate a slow connection through which a HTTP request
can be sent to the previous test server.

        delay := time.Duration(20) * time.Second
        conn := relay.NewProxy(delay, ts)

        client := &Client{Timeout: time.Duration(10) * time.Second}
        resp, _ := client.Get(conn.URL)

Note that the latency will double because it takes a round trip to and from the server.

Proxy can be placed in front of another proxy, and vice versa.

        delay := time.Duration(1) * time.Second
        ts := httptest.NewServer(http.HandlerFunc(handler))
        p3 := relay.NewProxy(delay, ts)
        p2 := relay.NewProxy(delay, p3)
        p1 := relay.NewProxy(delay, p2)
        resp, _ := client.Get(p1.URL)

Each hop to and from the target server will be delayed for one second.

Switcher works similarly to a proxy, except it "switches" between several backend servers
for each request in a round-robin fashion.

        ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       fmt.Fprint(w, "Hello world!")
        }
        ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       fmt.Fprint(w, "Hello mars!")
        }
        ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       fmt.Fprint(w, "Hello pluto!")
        }
        p := relay.NewProxy(delay, ts3)
        sw := relay.NewSwitcher(delay, []HTTPTestServer{ts1, ts2, p})

        resp1, _ := http.Get(sw.URL) // hits ts1
        resp2, _ := http.Get(sw.URL) // hits ts2
        resp3, _ := http.Get(sw.URL) // hits p, which eventually hits ts3
*/
package relay

