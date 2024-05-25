package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	net_url "net/url"
	"time"
)

func Current(w http.ResponseWriter, r *http.Request) {
	url := "http://api.weatherapi.com/v1/current.json?key=b8768196a632446db6e52729242505&q=bangkok"
	client := http.Client{
		Timeout: time.Second * 30,
	}

	kvGetUrl := net_url.URL{
		Host: "https://enhanced-seal-32374.upstash.io/get/bangkok/",
	}
	kvToken := "AX52ASQgNDU2ZDE5MTEtY2RlMC00MGI1LWFhMDYtZDRjNGNjYWI0OTA5MzgxYTU4MjQ1NWY1NDdiOTg4YWQ4NWUyZjMxM2M2OTQ="
	kvHeaders := http.Header{}
	kvHeaders.Add("Authorization", fmt.Sprintf("Bearer %s", kvToken))
	getCacheReq := http.Request{
		URL:    &kvGetUrl,
		Method: http.MethodGet,
		Header: kvHeaders,
	}
	cacheResp, err := client.Do(&getCacheReq)
	var respString string
	isFromCache := false
	if err != nil {
		fmt.Fprintf(w, `<h1>Get Cahce Error: %s</h1>`, err.Error())
	} else if cacheResp.StatusCode == http.StatusOK {
		b, err := io.ReadAll(cacheResp.Body)
		if err != nil {
			fmt.Fprintf(w, `<h1>Read Cache Body Error: %s</h1>`, err.Error())
		}

		type kvResponse struct {
			Result string `json:"result"`
		}
		var parsedKvResp kvResponse
		err = json.Unmarshal(b, &parsedKvResp)
		if err != nil {
			fmt.Fprintf(w, `<h1>Unmarshal Cache Body Error: %s</h1>`, err.Error())
		}

		if parsedKvResp.Result != "" {
			respString = parsedKvResp.Result
			isFromCache = true
		}
	}

	if !isFromCache {
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

		respString = string(b)
	}

	if respString == "" {
		//caching
		kvSetUrl := net_url.URL{
			Host: "https://enhanced-seal-32374.upstash.io/set/bangkok/",
			Path: respString,
		}
		cacheReq := http.Request{
			URL:    &kvSetUrl,
			Method: http.MethodPost,
			Header: kvHeaders,
		}
		_, err = client.Do(&cacheReq)
		if err != nil {
			fmt.Fprintf(w, `<h1>Create Cahce Error: %s</h1>`, err.Error())
		}
	}

	fmt.Fprintf(w, `<h1>Current Weather %v</h1><pre>%s<pre>`, isFromCache, respString)
}
