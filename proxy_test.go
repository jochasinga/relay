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

func TestBasicProxyConnection(t *testing.T) {
	
	Convey("GIVEN a back-end server", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(helloHandler))
		
		Convey("GIVEN a default front-end proxy", func() {
			latency := time.Duration(0) * time.Second
			proxy := NewProxy(latency, ts)
			
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

func TestBasicSwitcherConnection(t *testing.T) {
	
	Convey("GIVEN a few backend servers", t, func() {
		backends := []HTTPTestServer{
			httptest.NewServer(http.HandlerFunc(helloHandler)),
			httptest.NewServer(http.HandlerFunc(goodDayHandler)),
			httptest.NewServer(http.HandlerFunc(palomaHandler)),
		}
		
		Convey("GIVEN a frontend switcher", func() {
			latency := time.Duration(0) * time.Second
			sw := NewSwitcher(latency, backends)

			responses := []string{
				"Hello client!",
				"Good day client!",
				"Paloma client!",
			}

			Convey("WITH a first GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)

				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("EXPECT `Hello client!` from the first backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[0])
				})
			})

			Convey("WITH a second GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)

				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("EXPECT `Good day client!` from the second backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[1])
				})
			})

			Convey("WITH a third GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)

				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("EXPECT `Paloma client!` from the third backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[2])
				})
			})

			Convey("WITH a forth GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)

				Convey("EXPECT error to be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("EXPECT `Hello client!` from the first backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[2])
				})
			})
		})
		Reset(func() {
			for _, ts := range backends {
				ts.Close()
			}
		})
	})
}

