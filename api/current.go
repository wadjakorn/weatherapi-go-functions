package api

import (
	"bytes"
	"encoding/json"
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
	var respString string
	isFromCache := false
	kvToken := "AX52ASQgNDU2ZDE5MTEtY2RlMC00MGI1LWFhMDYtZDRjNGNjYWI0OTA5MzgxYTU4MjQ1NWY1NDdiOTg4YWQ4NWUyZjMxM2M2OTQ="
	getCacheReq, err := http.NewRequest(http.MethodGet, "https://enhanced-seal-32374.upstash.io/get/bangkok", nil)
	if err != nil {
		fmt.Printf(`Parse Get Cache URL Error: %s`, err.Error())
	} else {
		getCacheReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kvToken))
		cacheResp, err := client.Do(getCacheReq)

		if err != nil {
			fmt.Printf(`Get Cahce Error: %s`, err.Error())
		} else if cacheResp.StatusCode == http.StatusOK {
			b, err := io.ReadAll(cacheResp.Body)
			if err != nil {
				fmt.Printf(`Read Cache Body Error: %s`, err.Error())
			}

			type kvResponse struct {
				Result string `json:"result"`
			}
			var parsedKvResp kvResponse
			err = json.Unmarshal(b, &parsedKvResp)
			if err != nil {
				fmt.Printf(`Unmarshal Cache Body Error: %s`, err.Error())
			}

			if parsedKvResp.Result != "" {
				respString = parsedKvResp.Result
				isFromCache = true
			}
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

		//caching
		kvSetBody := bytes.NewReader([]byte(fmt.Sprintf(`["SET", "bangkok", "%s", "ex" ,"10"]`, respString)))
		kvSeReq, err := http.NewRequest(http.MethodPost, "https://enhanced-seal-32374.upstash.io", kvSetBody)
		if err != nil {
			fmt.Printf(`Parse Set Cache URL Error: %s`, err.Error())
		} else {
			kvSeReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kvToken))
			_, err = client.Do(kvSeReq)
			if err != nil {
				fmt.Printf(`Create Cahce Error: %s`, err.Error())
			}
		}
	}

	fmt.Fprintf(w, `<h1>Current Weather</h1><br>
	<p>cache: %v</p><br>
	<pre>%s</pre>`, isFromCache, respString)
}
