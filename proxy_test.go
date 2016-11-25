package relay

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)

var helloHandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello client!")
})

func TestStartingAndClosingProxy(t *testing.T) {
	delay := time.Duration(0)
	backend := httptest.NewServer(helloHandlerFunc)

	proxy := NewUnstartedProxy(delay, backend)
	proxy.Start()

	resp, _ := http.Get(proxy.URL)
	assert.NotNil(t, resp, "Response should not be empty")

	proxy.Close()
	_, err := http.Get(proxy.URL)
	assert.NotNil(t, err, "Error should not be empty")
}

func TestProxyConnection(t *testing.T) {
	assert := assert.New(t)

	backend := httptest.NewServer(helloHandlerFunc)
	latency := time.Duration(0)
	
	proxy := NewProxy(latency, backend)
	defer proxy.Close()

	resp, err := http.Get(proxy.URL)
	assert.Nil(err, "Error should be empty")

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err, "Error should be empty")

	assert.Equal("Hello client!", string(body), "Response text should be \"Hello client!\"")
}

func TestUnstartedProxyConnection(t *testing.T) {
	assert := assert.New(t)
	
	backend := httptest.NewServer(helloHandlerFunc)
	latency := time.Duration(0) 
	proxy := NewUnstartedProxy(latency, backend)
	proxy.Start()
	defer proxy.Close()

	resp, err := http.Get(proxy.URL)
	assert.Nil(err, "Error should be empty")

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("Hello client!", string(body), "Response text should be \"Hello client!\"")
}

func TestMultipleProxies(t *testing.T) {
	assert := assert.New(t)
	
	backend := httptest.NewServer(helloHandlerFunc)
	latency := time.Duration(0) * time.Second
	proxy3 := NewProxy(latency, backend)
	proxy2 := NewProxy(latency, proxy3)
	proxy1 := NewProxy(latency, proxy2)

	defer func() {
		proxy3.Close()
		proxy2.Close()
		proxy1.Close()
	}()

	// Send request to the front-most proxy
	resp, err := http.Get(proxy1.URL)
	assert.Nil(err, "Error should be empty")
	
	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal("Hello client!", string(body), "Response text should be \"Hello client!\"")
}



