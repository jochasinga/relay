relay
=====

[![GoDoc](https://godoc.org/github.com/jochasinga/relay?status.svg)](https://godoc.org/github.com/jochasinga/relay)  [![Build Status](https://drone.io/github.com/jochasinga/relay/status.png)](https://drone.io/github.com/jochasinga/relay/latest)  [![Coverage Status](https://coveralls.io/repos/github/jochasinga/relay/badge.svg?branch=master)](https://coveralls.io/github/jochasinga/relay?branch=master)

Powered up Go HTTP Server for comprehensive end-to-end HTTP tests.

**Relay** consists of two components, `Proxy` and `Switcher`. They are both
HTTP [middlewares](https://justinas.org/writing-http-middleware-in-go/) which 
wrap around the target server's handler to simulate latent connections, proxy 
servers, or load balancers.

Proxy
-----

A `Proxy` is used to place in front of any `HTTPTestServer` (`httptest.Server`,
`Proxy`, or `Switcher`) to simulate a proxy server or a connection with some
network latency. It takes a latency unit in `time.Duration` and a backend 
`HTTPTestServer` as arguments.

Let's begin setting up a basic `httptest.Server` to send a request to.

```go

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

```

Now let's simulate a slow connection through which an HTTP request
can be sent to the previous test server with `Proxy`.

```go

delay := time.Duration(20) * time.Second
conn := relay.NewProxy(delay, ts)
client := &Client{Timeout: time.Duration(10) * time.Second}
resp, _ := client.Get(conn.URL)

```
Note that the latency will double because it takes a round trip to and 
from the server.

`Proxy` can be placed in front of another proxy, and vice versa.

```go

delay := time.Duration(1) * time.Second
ts := httptest.NewServer(http.HandlerFunc(handler))
p3 := relay.NewProxy(delay, ts)
p2 := relay.NewProxy(delay, p3)
p1 := relay.NewProxy(delay, p2)
resp, _ := client.Get(p1.URL)

```

In the above code, Each hop to and from the target server `ts` will be delayed 
for one second, total to six seconds of latency.

Switcher
--------

`Switcher` works similarly to a proxy, except with each request it "switches" between several backend servers in a round-robin fashion.

```go

ts1 := httptest.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    fmt.Fprint(w, "Hello world!")
}))
ts2 := httptest.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    fmt.Fprint(w, "Hello mars!")
}))
ts3 := httptest.New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    fmt.Fprint(w, "Hello pluto!")
}))

// a proxy sitting in front of ts3
p := relay.NewProxy(delay, ts3)
sw := relay.NewSwitcher(delay, []HTTPTestServer{ts1, ts2, p})

resp1, _ := http.Get(sw.URL) // hits ts1
resp2, _ := http.Get(sw.URL) // hits ts2
resp3, _ := http.Get(sw.URL) // hits p, which eventually hits ts3
resp4, _ := http.Get(sw.URL) // hits ts1

```

TODO
----
+ Make `Proxy` a standalone `httptest.Server` with optional `backend=nil`.
+ Add other common middlewares to each proxy and switcher.
+ Add options to inject middleware into each proxy and switcher.
+ Add more "load-balancing" capabilities to `Switcher`.


