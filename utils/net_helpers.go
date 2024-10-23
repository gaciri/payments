package utils

import "net/url"

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
