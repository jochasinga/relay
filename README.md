relay
=====

[![Build Status](https://drone.io/github.com/jochasinga/relay/status.png)](https://drone.io/github.com/jochasinga/relay/latest)  [![Build Status](https://drone.io/github.com/jochasinga/relay/status.png)](https://drone.io/github.com/jochasinga/relay/latest)

Powered up Go httptest.Server for comprehensive end-to-end HTTP tests.

Relay consists of two components, `Proxy` and `Switcher`. They are both
HTTP middlewares which wrap around the target server's handler to simulate
latent connections, proxy servers, or load balancers.

A Proxy is used to place in front of any HTTPTestServer (httptest.Server,
Proxy, or Switcher) to simulate a proxy server or a connection with some
network, I/O or CPU latency. It takes a latency unit in time.Duration and
a backend HTTPTestServer as arguments.

Let's begin setting up a basic httptest.Server to test against.

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

Now let's simulate a slow connection through which a HTTP request
can be sent to the previous test server.

```go

delay := time.Duration(20) * time.Second
conn := relay.NewProxy(delay, ts)
client := &Client{Timeout: time.Duration(10) * time.Second}
resp, _ := client.Get(conn.URL)

```
Note that the latency will double because it takes a round trip to and 
from the server.

Proxy can be placed in front of another proxy, and vice versa.

```go

delay := time.Duration(1) * time.Second
ts := httptest.NewServer(http.HandlerFunc(handler))
p3 := relay.NewProxy(delay, ts)
p2 := relay.NewProxy(delay, p3)
p1 := relay.NewProxy(delay, p2)
resp, _ := client.Get(p1.URL)

```

Each hope to and from the target server will be delayed for one second.

Switcher works similarly to a proxy, except it "switches" between several backend servers
for each request in a round-robin fashion.

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
p := relay.NewProxy(delay, ts3)
sw := relay.NewSwitcher(delay, []HTTPTestServer{ts1, ts2, p}

resp1, _ := http.Get(sw.URL) // hits ts1
resp2, _ := http.Get(sw.URL) // hits ts2
resp3, _ := http.Get(sw.URL) // hits p, which eventually hits ts3

```

TODO
----
+ Making `Proxy` a standalone `httptest.Server` with optional `backend=nil`.
+ Add options to inject middleware into each proxy and switcher.


