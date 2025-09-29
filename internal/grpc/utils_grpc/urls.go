package utils

import (
	"fmt"
	"net/url"
)

const (
	NURL = "nurl"
	BURL = "burl"
)

func WrapURL(hostname, originalURL, globalId, isItNurlOrBurl string) string {
	encodedURL := url.QueryEscape(originalURL)
	return fmt.Sprintf("https://%s/%s?id=%s&url=%s",
		hostname, isItNurlOrBurl, globalId, encodedURL)
}
