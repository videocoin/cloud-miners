package rpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ipwhoisResponse struct {
	IP        string  `json:"ip"`
	Success   bool    `json:"success"`
	Latitude  float32 `json:"latitude,string"`
	Longitude float32 `json:"longitude,string"`
}

func GetGeoLocation(ip string) (float32, float32, error) {
	url := fmt.Sprintf("http://free.ipwhois.io/json/%s", ip)
	client := http.Client{Timeout: time.Second * 2}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, err
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	if res.StatusCode != 200 {
		return 0, 0, fmt.Errorf("Ip Whois service failure. Bad status code: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}

	geoResponse := ipwhoisResponse{}
	err = json.Unmarshal(body, &geoResponse)
	if err != nil {
		return 0, 0, err
	}
	if !geoResponse.Success {
		return 0, 0, fmt.Errorf("Ip Whois service failure. Bad response: %s", body)
	}

	return geoResponse.Latitude, geoResponse.Longitude, nil
}
