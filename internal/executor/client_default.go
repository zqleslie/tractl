//go:build !js

package executor

import (
	"net/http"
	"time"
)

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 0,
		Transport: &http.Transport{
			DisableKeepAlives: false,
			MaxIdleConns:      100,
			IdleConnTimeout:   90 * time.Second,
		},
		CheckRedirect: redirectRecorder,
	}
}
