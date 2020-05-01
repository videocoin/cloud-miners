package rpc

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

func GetGeoLocation(ip string) (float64, float64, error) {
	db, err := geoip2.Open("./GeoLite/GeoLite2.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// If you are using strings that may be invalid, check that ip is not nil
	parsedIP := net.ParseIP(ip)
	record, err := db.City(parsedIP)
	if err != nil {
		log.Fatal(err)
	}

	return record.Location.Latitude, record.Location.Longitude, nil
}
