package server

import (
	"fmt"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	handleFunc := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/path1", handleFunc)
	mux.HandleFunc("/path2", handleFunc)

	s := NewServer()
	s.Server().Addr = ":8888"
	s.Server().Handler = mux

	go func() {
		s.Start(false)
	}()

	testRequest := func(path string, tsp *http.Transport, status int) {
		req, err := http.NewRequest("GET", path, nil)

		if err != nil {
			t.Error("can't create request to path", path)
		}

		if resp, err := tsp.RoundTrip(req); err != nil {
			t.Error("request", path, "failed", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == status {
				t.Log("Success request to", path, "ok", "status", resp.StatusCode, "==", status)
			} else {
				t.Error("Bad response status. Need", status, "responsed", resp.StatusCode)
			}
		}
	}

	tsp := &http.Transport{}

	testRequest("http://localhost:8888/path1", tsp, http.StatusOK)
	testRequest("http://localhost:8888/path2", tsp, http.StatusOK)

	testRequest("http://localhost:8888/", tsp, http.StatusNotFound)
	testRequest("http://localhost:8888/path3", tsp, http.StatusNotFound)
	testRequest("http://localhost:8888/path4", tsp, http.StatusNotFound)

	if err := s.Stop(); err != nil {
		t.Error("Stop server gives error", err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8888/path4", nil)

	if err != nil {
		t.Error("can't create request", err)
	}

	if resp, err := tsp.RoundTrip(req); err != nil {
		if err.Error() == "dial tcp 127.0.0.1:8888: connection refused" {
			t.Log("Stoped server don't response")
		} else {
			t.Error("Some different error!", err)
		}
	} else {
		defer resp.Body.Close()
		t.Error("Stopped server responsed!")
	}
}
