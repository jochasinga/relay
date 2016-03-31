package robin

import (
	"io/ioutil"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	helloHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello client!")
	}
	goodDayHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Good day client!")
	}
	palomaHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Paloma client!")
	}
)

func TestBasicConnection(t *testing.T) {
	
	Convey("GIVEN a back-end server", t, func() {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(helloHandler))
		
		Convey("GIVEN a default front-end proxy", func() {
			latency := time.Duration(1) * time.Second
			proxy := NewProxy(latency, ts)
			
			Convey("WITH a basic GET request to the front-end proxy", func() {
				resp, err := http.Get(proxy.URL)
				
				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})
				
				Convey("EXPECT response to be correct", func() {
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
			latency := time.Duration(1) * time.Second
			proxy := NewUnstartedProxy(latency, ts)
			proxy.SetPort("8888")
			proxy.Start()

			Convey("WITH a basic GET request to the front-end proxy", func() {
				resp, err := http.Get(proxy.URL)
				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("EXPECT response to be correct", func() {
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
			latency := time.Duration(1) * time.Second
			proxy3 := NewProxy(latency, ts)
			proxy2 := NewProxy(latency, proxy3)
			proxy1 := NewProxy(latency, proxy2)

			Convey("WITH a basic GET request to the first", func() {
				resp, err := http.Get(proxy1.URL)
				
				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})

				Convey("EXPECT response to be correct", func() {
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
		Convey("GIVEN a few more backend servers", func() {
			ts1 := httptest.NewServer(http.HandlerFunc(goodDayHandler))
			ts2 := httptest.NewServer(http.HandlerFunc(palomaHandler))
			
			Convey("GIVEN a frontend balancer", func() {
				backends := []HTTPTestServer{ts, ts1, ts2}
				b := NewUnstartedBalancer(backends)
				responses := []string{
					"Hello client!",
					"Good day client!",
					"Paloma client!",
				}

				for _, r := range responses {
					Convey("WITH a basic GET request", func() {
						resp, err := http.Get(b.URL)

						Convey("EXPECT error to be nil", func() {
							So(err, ShouldBeNil)
						})
						Convey("EXPECT response to be correct", func() {
							b, err := ioutil.ReadAll(resp.Body)
							if err != nil {
								t.Error(err)
							}
							So(string(b), ShouldEqual, r)
						})
					})
				}

			})
			Reset(func() {
				ts1.Close()
				ts2.Close()
			})
		})
		Reset(func() {
			ts.Close()
		})
	})
}

