package statc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func newHTTPTest(t *testing.T) (*check.C, *sTest, *HTTPMuxer) {
	c, st := newTest(t, &Config{})
	mux := st.NewHTTPMuxer("http")
	return c, st, mux
}

func TestHTTPBasic(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	mux.HandlerFunc("GET", "/sleep/1",
		func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond)
		})
	mux.HandlerFunc("GET", "/sleep/5",
		func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond * 5)
		})
	mux.HandlerFunc("GET", "/404",
		func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "404", 404)
		})
	mux.HandlerFunc("GET", "/500",
		func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "500", 500)
		})
	mux.HandlerFunc("GET", "/164",
		func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "164", 164)
		})

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	wg := sync.WaitGroup{}
	get := func(url string) {
		defer wg.Done()

		resp, err := http.Get(url)
		c.MustNotError(err)
		resp.Body.Close()
	}

	for i := 0; i < 10; i++ {
		wg.Add(5)
		go get(fmt.Sprintf("%s/sleep/1/", srv.URL))
		go get(fmt.Sprintf("%s/sleep/5/", srv.URL))
		go get(fmt.Sprintf("%s/404", srv.URL))
		go get(fmt.Sprintf("%s/500", srv.URL))
		go get(fmt.Sprintf("%s/164", srv.URL))
	}

	wg.Wait()

	snap := st.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name, st.Val)
	}

	c.Equal(snap.Get("http./sleep/1.GET.200.count").Val.(int64), 10)
	c.Equal(snap.Get("http./500.GET.500.count").Val.(int64), 10)
	c.Equal(snap.Get("http./404.GET.404.count").Val.(int64), 10)
	c.Equal(snap.Get("http./164.GET.0.count").Val.(int64), 10)
}

func TestHTTPPanic(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	mux.HandlerFunc("GET", "/panic",
		func(rw http.ResponseWriter, req *http.Request) {
			panic("i give up")
		})

	h, p, _ := mux.R.Lookup("GET", "/panic")
	c.MustTrue(h != nil)

	c.Panic(func() {
		rw := httptest.NewRecorder()
		h(rw, nil, p)
	})

	snap := st.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name, st.Val)
	}

	c.Equal(snap.Get("http./panic.GET.panic.count").Val.(int64), 1)
}

func TestHTTPStatusHandler(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	st.NewTimer("some.timer", 100).Add(time.Second)
	st.NewCounter("module.counter", false).Add(100)
	st.NewGauge("my.gauge").Set(9)
	st.NewStringGauge("str.gauge").Set("some string")
	st.doSnapshot()

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/_status")
	c.MustNotError(err)
	defer resp.Body.Close()

	r, err := ioutil.ReadAll(resp.Body)
	c.MustNotError(err)

	out := `{
	"module": {
		"counter": 100
	},
	"my": {
		"gauge": 9
	},
	"some": {
		"timer": {
			"count": 1,
			"max": 1000000000,
			"mean": 1000000000,
			"min": 1000000000,
			"p50": 1000000000,
			"p75": 1000000000,
			"p90": 1000000000,
			"p95": 1000000000,
			"stddev": 0
		}
	},
	"str": {
		"gauge": "some string"
	}` + "\n}"

	c.Equal(string(r), out)
}

func TestHTTPStatusHandlerError(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	st.lastSnap = Snapshot{
		Stat{
			Name: "blah",
			Val:  nil,
		},
	}

	resp, err := http.Get(srv.URL + "/_status")
	c.MustNotError(err)
	defer resp.Body.Close()
	c.Equal(resp.StatusCode, http.StatusInternalServerError)
}