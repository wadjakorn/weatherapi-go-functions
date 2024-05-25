package api

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func Current(w http.ResponseWriter, r *http.Request) {
	url := "http://api.weatherapi.com/v1/current.json?key=b8768196a632446db6e52729242505&q=bangkok"
	client := http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(w, `<h1>Error: %s</h1>`, err.Error())
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, `<h1>StatusCode: %d</h1>`, resp.StatusCode)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w, `<h1>Read Body Error: %s</h1>`, err.Error())
		return
	}

	fmt.Fprintf(w, `<h1>Current Weather</h1><pre>%s<pre>`, string(b))
}
