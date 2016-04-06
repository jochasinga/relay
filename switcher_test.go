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

var (
	helloMarsHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello Mars client!")
	}
	goodDayHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Good day client!")
	}
	palomaHandler = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Paloma client!")
	}
)

func TestBasicSwitcherUtility(t *testing.T) {
	Convey("GIVEN an unstarted switcher", t, func() {
		delay := time.Duration(0)
		ts := httptest.NewServer(http.HandlerFunc(helloHandler))
		sw := NewUnstartedSwitcher(delay, []HTTPTestServer{ ts })
		Convey("WITH a call to `Start()`", func() {
			sw.Start()
			Convey("EXPECT proxy to be running", func() {
				resp, _ := http.Get(sw.URL)
				So(resp, ShouldNotBeNil)
			})
		})
		Convey("WITH a call to `Close()`", func() {
			sw.Close()
			Convey("EXPECT proxy to be closed", func() {
				_, err := http.Get(sw.URL)
				So(err, ShouldNotBeNil)
			})
		})
		Reset(func() {
			ts.Close()
			sw.Close()
		})
	})
}

func TestBasicSwitcherConnection(t *testing.T) {
	Convey("GIVEN a few backend servers", t, func() {
		backends := []HTTPTestServer{
			httptest.NewServer(http.HandlerFunc(helloMarsHandler)),
			httptest.NewServer(http.HandlerFunc(goodDayHandler)),
			httptest.NewServer(http.HandlerFunc(palomaHandler)),
		}
		
		Convey("GIVEN a frontend switcher", func() {
			latency := time.Duration(0) * time.Second
			sw := NewSwitcher(latency, backends)
			responses := []string{
				"Hello Mars client!",
				"Good day client!",
				"Paloma client!",
			}
			Convey("WITH a first GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)
				// TODO: Handling error here instead of asserting it.
				if err != nil {
					t.Error(err)
				}
				// TODO: This Convey actually sends another request to the switcher,
				//       making the test results inaccurate.
				//
				// Convey("EXPECT error to be nil", func() {
				//         So(err, ShouldBeNil)
				//})
				//
				Convey("EXPECT `Hello Mars client!` from the first backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[0])
				})
			})
			Convey("WITH a second GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)
				// TODO: Handling error here instead of asserting it.
				if err != nil {
					t.Error(err)
				}
				// TODO: This Convey actually sends another request to the switcher,
				//       making the test results inaccurate.
				//
				// Convey("EXPECT error to be nil", func() {
				//         So(err, ShouldBeNil)
				// })
                                //
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
				// TODO: Handling error here instead of asserting it.
				if err != nil {
					t.Error(err)
				}
				// TODO: This Convey actually sends another request to the switcher,
				//       making the test results inaccurate.
				//
				// Convey("EXPECT error to be nil", func() {
				//         So(err, ShouldBeNil)
				// })
				//
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
				// TODO: Handling error here instead of asserting it.
				if err != nil {
					t.Error(err)
				}
				// TODO: This Convey actually sends another request to the switcher,
				//       making the test results inaccurate.
				//
				// Convey("EXPECT error to be nil", func() {
				//         So(err, ShouldBeNil)
				// })
				//
				Convey("EXPECT `Hello Mars client!` from the first backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[0])
				})
			})
			Convey("WITH a fifth GET request to the switcher", func() {
				resp, err := http.Get(sw.URL)
				// TODO: Handling error here instead of asserting it.
				if err != nil {
					t.Error(err)
				}
				// TODO: This Convey actually sends another request to the switcher,
				//       making the test results inaccurate.
				//
				// Convey("EXPECT error to be nil", func() {
				//         So(err, ShouldBeNil)
				// })
				//
				Convey("EXPECT `Good day client!` from the second backend server", func() {
					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						t.Error(err)
					}
					So(string(b), ShouldEqual, responses[1])
				})
			})
			Reset(func() {
				sw.Close()
			})
		})
		Reset(func() {
			for _, ts := range backends {
				ts.Close()
			}
		})
	})
}

