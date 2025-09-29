package geoBadIp

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type GeoIPRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

type GeoIPService struct {
	db *maxminddb.Reader
}

func NewGeoIPService(dbPath string) (*GeoIPService, error) {
	reader, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &GeoIPService{db: reader}, nil
}

func (g *GeoIPService) Close() {
	if g.db != nil {
		g.db.Close()
	}
}

func (g *GeoIPService) GetCountryISO(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf(
			"%w %s",
			BadIpFormatError,
			ip,
		)
	}
	var rec GeoIPRecord
	if err := g.db.Lookup(ip, &rec); err != nil {
		return "", fmt.Errorf(
			"%w %s: %w",
			InnerLookupIpError,
			ip,
			err,
		)
	}
	return rec.Country.ISOCode, nil
}
