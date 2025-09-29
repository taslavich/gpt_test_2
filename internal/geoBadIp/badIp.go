package geoBadIp

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type Record struct {
	IsAnonymous       bool `maxminddb:"is_anonymous"`
	IsPublicProxy     bool `maxminddb:"is_public_proxy"`
	IsTorExitNode     bool `maxminddb:"is_tor_exit_node"`
	IsHostingProvider bool `maxminddb:"is_hosting_provider"`
	IsVPN             bool `maxminddb:"is_anonymous_vpn"`
	IsResidentialVPN  bool `maxminddb:"is_residential_proxy"`
}

type BadIPService struct {
	db *maxminddb.Reader
}

func NewBadIPService(dbPath string) (*BadIPService, error) {
	reader, err := maxminddb.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &BadIPService{db: reader}, nil
}

func (s *BadIPService) Close() { s.db.Close() }

func (s *BadIPService) IsBad(ipStr string) (bool, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return true, fmt.Errorf(
			"%w %s",
			BadIpFormatError,
			ip,
		)
	}
	var rec Record
	if err := s.db.Lookup(ip, &rec); err != nil {
		return false, fmt.Errorf(
			"%w %s: %w",
			InnerLookupIpError,
			ip,
			err,
		)
	}
	switch {
	case rec.IsTorExitNode:
		return true, fmt.Errorf(
			"%w %s",
			TorExitError,
			ip,
		)
	case rec.IsPublicProxy:
		return true, fmt.Errorf(
			"%w %s",
			PublicProxyError,
			ip,
		)
	case rec.IsAnonymous || rec.IsVPN || rec.IsHostingProvider || rec.IsResidentialVPN:
		return true, fmt.Errorf(
			"%w %s",
			AnonymousIpError,
			ip,
		)
	default:
		return false, nil
	}
}
