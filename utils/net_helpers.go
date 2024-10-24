package utils

import (
	"net/url"
	"strings"
)

func JoinUrlPaths(base, path string) (string, error) {
	baseUrl, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	baseUrl.Path, err = url.JoinPath(baseUrl.Path, path)
	if err != nil {
		return "", err
	}
	return baseUrl.String(), nil
}
func ExtractDomain(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	host := parsedUrl.Host

	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	return host, nil
}
