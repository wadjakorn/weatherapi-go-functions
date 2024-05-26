package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CustomResponse struct {
	Result string   `json:"result"`
	Errors []string `json:"errors"`
	Cached bool     `json:"cached"`
}

func Bangkok(w http.ResponseWriter, r *http.Request) {
	q := "bangkok"
	url := "http://api.weatherapi.com/v1/current.json?key=b8768196a632446db6e52729242505&q=bangkok"
	client := http.Client{
		Timeout: time.Second * 30,
	}
	cresp := CustomResponse{
		Cached: false,
		Errors: []string{},
	}

	defer func() {
		b, _ := json.Marshal(cresp)
		fmt.Fprintf(w, "%s", string(b))
	}()

	kvToken := "AX52ASQgNDU2ZDE5MTEtY2RlMC00MGI1LWFhMDYtZDRjNGNjYWI0OTA5MzgxYTU4MjQ1NWY1NDdiOTg4YWQ4NWUyZjMxM2M2OTQ="
	getCacheReq, err := http.NewRequest(http.MethodGet, "https://enhanced-seal-32374.upstash.io/get/"+q, nil)
	if err != nil {
		cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Parse Get Cache URL Error: %s`, err.Error()))
	} else {
		getCacheReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kvToken))
		cacheResp, err := client.Do(getCacheReq)

		if err != nil {
			fmt.Printf(`Get Cahce Error: %s`, err.Error())
		} else if cacheResp.StatusCode == http.StatusOK {
			b, err := io.ReadAll(cacheResp.Body)
			if err != nil {
				cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Read Cache Body Error: %s`, err.Error()))
			} else {
				type kvResponse struct {
					Result string `json:"result"`
				}
				var parsedKvResp kvResponse
				err = json.Unmarshal(b, &parsedKvResp)
				if err != nil {
					cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Unmarshal Cache Body Error: %s`, err.Error()))
				}

				if parsedKvResp.Result != "" {
					cresp.Result = parsedKvResp.Result
					cresp.Cached = true
				}
			}
		}
	}

	if !cresp.Cached {
		resp, err := client.Get(url)
		if err != nil {
			cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Client Get Error: %s`, err.Error()))
			return
		}

		if resp.StatusCode != http.StatusOK {
			cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Client Get StatusCode: %d`, resp.StatusCode))
			return
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Read Body Error: %s`, err.Error()))
			return
		}

		cresp.Result = string(b)
		kvValue := strings.ReplaceAll(cresp.Result, `"`, `\"`)

		//caching
		kvSetBody := bytes.NewReader([]byte(fmt.Sprintf(`["SET", "%s", "%s", "ex" ,"60"]`, q, kvValue)))
		kvSeReq, err := http.NewRequest(http.MethodPost, "https://enhanced-seal-32374.upstash.io", kvSetBody)
		if err != nil {
			cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Parse Set Cache URL Error: %s`, err.Error()))
		} else {
			kvSeReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", kvToken))
			_, err = client.Do(kvSeReq)
			if err != nil {
				cresp.Errors = append(cresp.Errors, fmt.Sprintf(`Create Cahce Error: %s`, err.Error()))
			}
		}
	}
}
