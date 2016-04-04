package relay

import (
	"io/ioutil"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

var helloHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello client!")
}

func TestBasicProxyConnection(t *testing.T) {
	
	Convey("GIVEN a back-end server", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(helloHandler))
		
		Convey("GIVEN a default front-end proxy", func() {
			latency := time.Duration(0) * time.Second
			proxy := NewProxy(latency, ts)
			
			Convey("WITH a basic GET request to the front-end proxy", func() {
				resp, err := http.Get(proxy.URL)
				if err != nil {
					t.Error(err)
				}

				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})
				
				Convey("EXPECT response to be `Hello client!`", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, "Hello client!")
				})
			})
			Reset(func() {
				proxy.Close()
			})
		})
		Convey("GIVEN a front-proxy with a set port", func() {
			latency := time.Duration(0) * time.Second
			proxy := NewUnstartedProxy(latency, ts)
			proxy.SetPort("8888")
			proxy.Start()

			Convey("WITH a basic GET request to the front-end proxy", func() {
				resp, err := http.Get(proxy.URL)
				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("EXPECT response to be `Hello client!`", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, "Hello client!")
				})
			})
			Reset(func() {
				proxy.Close()
			})
		})
		Convey("GIVEN several proxies put in front", func() {
			latency := time.Duration(0) * time.Second
			proxy3 := NewProxy(latency, ts)
			proxy2 := NewProxy(latency, proxy3)
			proxy1 := NewProxy(latency, proxy2)

			Convey("WITH a basic GET request to the front-most proxy", func() {
				resp, err := http.Get(proxy1.URL)
				
				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("EXPECT response to be `Hello client!`", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, "Hello client!")
				})
			})
			Reset(func() {
				proxy1.Close()
				proxy2.Close()
				proxy3.Close()
			})
		})
		Reset(func() {
			ts.Close()
		})
	})
}



