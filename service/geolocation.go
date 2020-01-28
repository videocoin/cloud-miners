package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ipwhoisResponse struct {
	Ip        string  `json:"ip"`
	Success   bool    `json:"success"`
	Latitude  float32 `json:"latitude,string"`
	Longitude float32 `json:"longitude,string"`
}

var (
	ErrIpwhoisFailure = errors.New("Ip Whois service failure")
)

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
		return 0, 0, ErrIpwhoisFailure
	}

	return geoResponse.Latitude, geoResponse.Longitude, nil
}
