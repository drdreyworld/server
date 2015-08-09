package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	transport := &http.Transport{MaxIdleConnsPerHost: 1}
	request, err := http.NewRequest("GET", "http://127.0.0.1:8000/", nil)

	if err != nil {
		log.Fatalln(err.Error())
	}

	for {
		resp, err := transport.RoundTrip(request)

		if err != nil {
			print("E")
		} else {
			resp.Body.Close()
			print(".")
		}
		time.Sleep(time.Millisecond)
	}
}
