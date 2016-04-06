relay
=====

[![GoDoc](https://godoc.org/github.com/jochasinga/relay?status.svg)](https://godoc.org/github.com/jochasinga/relay)  [![Build Status](https://drone.io/github.com/jochasinga/relay/status.png)](https://drone.io/github.com/jochasinga/relay/latest)  [![Coverage Status](https://coveralls.io/repos/github/jochasinga/relay/badge.svg?branch=master)](https://coveralls.io/github/jochasinga/relay?branch=master)  [![Flattr this git repo](http://api.flattr.com/button/flattr-badge-large.png)](https://flattr.com/submit/auto?user_id=jochasinga&url=https://github.com/jochasinga/relay&title=Relay&language=English&tags=github&category=software)

Powered up Go HTTP Server for comprehensive end-to-end HTTP tests.

**Relay** consists of two components, `Proxy` and `Switcher`. They are both
HTTP [middlewares](https://justinas.org/writing-http-middleware-in-go/) which 
wrap around the target server's handler to simulate latent connections, proxy 
servers, or load balancers.

Usage
-----
To use `relay` in your test, install with 

```bash

$ go get github.com/jochasinga/relay

```

Or better, use a [package manager](https://github.com/golang/go/wiki/PackageManagementTools) like [Godep](https://github.com/tools/godep) or [Glide](https://glide.sh/).

HTTPTestServer
--------------
Every instance, including the standard `httptest.Server`, implements `HTTPTestServer` 
interface, which means that they can be used interchangeably as a general "server".

Proxy
-----
A `Proxy` is used to place in front of any `HTTPTestServer` (`httptest.Server`,
`Proxy`, or `Switcher`) to simulate a proxy server or a connection with some
network latency. It takes a latency unit in `time.Duration` and a backend 
`HTTPTestServer` as arguments.

Switcher
--------
`Switcher` works similarly to `Proxy`, except with each request it "switches" between 
several backend `HTTPTestServer` in a round-robin fashion.

Examples
--------
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

Now let's use `Proxy` to simulate a slow connection through which an HTTP request
can be sent to test server.

```go

// Connection takes 4s to and from the test server.
delay := time.Duration(2) * time.Second
// Client takes 3s before timing out.
timeout := time.Duration(3) * time.Second
// Create a new proxy with the delay and test server backend.
conn := relay.NewProxy(delay, ts)
client := &Client{ Timeout: timeout }
start := time.Now()
_, _ := client.Get(conn.URL)
elapsed := time.Since(start)
deviation := time.Duration(50) * time.Millisecond

if elapsed >= timeout + deviation {
	    t.Error("Client took too long to time out.")
}

```
Note that the latency will double because of the round trip to and 
from the server.

`Proxy` can be placed in front of another proxy, and vice versa. So you 
can create a chain of test proxies this way:

```go

delay := time.Duration(1) * time.Second
ts := httptest.NewServer(http.HandlerFunc(handler))
p3 := relay.NewProxy(delay, ts)
p2 := relay.NewProxy(delay, p3)
p1 := relay.NewProxy(delay, p2)
start := time.Now()
_, _ := client.Get(p1.URL)
elapsed := time.Since(start)
deviation := time.Duration(50) * time.Millisecond
if elapsed >= (time.Duration * 6) + deviation {
	    t.Error("Client took longer than expected.")
}

```

In the above code, Each hop to and from the target server `ts` will be delayed 
for one second, total to six seconds of latency.

`Switcher` can be used instead of `Proxy` to simulate a round-robin load-balancing proxy
or just to switch between several test servers' handlers for convenience.

```go

ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    fmt.Fprint(w, "Hello world!")
}))
ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    fmt.Fprint(w, "Hello mars!")
}))
ts3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    fmt.Fprint(w, "Hello pluto!")
}))

// a proxy sitting in front of ts3
p := relay.NewProxy(delay, ts3)
sw := relay.NewSwitcher(delay, []HTTPTestServer{ts1, ts2, p})

resp1, _ := http.Get(sw.URL) // hits ts1
resp2, _ := http.Get(sw.URL) // hits ts2
resp3, _ := http.Get(sw.URL) // hits p, which eventually hits ts3
resp4, _ := http.Get(sw.URL) // hits ts1 again, and so on.

```

Also please check out this [introduction to writing test with goconvey and relay on Medium](https://medium.com/code-zen/go-http-test-with-relay-deade218fd3d#.ka0a2x19z).

TODO
----
+ Make `Proxy` a standalone `httptest.Server` with optional `backend=nil`.
+ Add other common middlewares to each proxy and switcher.
+ Add options to inject middleware into each proxy and switcher.
+ Add more "load-balancing" capabilities to `Switcher`.

CONTRIBUTE
----------
Please feel free to open an issue or send a pull request. 
Please see [goconvey](https://github.com/smartystreets/goconvey) on how to write BDD tests for relay.
Contact me on twitter [@jochasinga](http://twitter.com).
Fuel me with high-quality caffeine to continue working on cool code -> [![Flattr this git repo](http://api.flattr.com/button/flattr-badge-large.png)](https://flattr.com/submit/auto?user_id=jochasinga&url=https://github.com/jochasinga/relay&title=Relay&language=English&tags=github&category=software)
