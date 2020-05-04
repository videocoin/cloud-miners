package rpc

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

func GetLatLon(ip string) (float64, float64, error) {
	db, err := geoip2.Open("/data/GeoLite2.mmdb")
	if err != nil {
		return 0, 0, err
	}
	defer db.Close()

	parsedIP := net.ParseIP(ip)
	record, err := db.City(parsedIP)
	if err != nil {
		return 0, 0, err
	}

	return record.Location.Latitude, record.Location.Longitude, nil
}
