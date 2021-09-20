package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func forward(port int, target string, maxRetries int) error {
	if target == "" {
		log.Fatalf("missing target\n")
	}
	if port <= 0 {
		return fmt.Errorf("invalid port %d\n", port)
	}
	targetURL, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("error parsing URL: %v\n", err)
	}
	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	forwardRequestHandler := newRequestForwarder(port, targetURL, httpClient)
	portStr := strconv.Itoa(port)

	retryCount := 0
	for retryCount < maxRetries {
		log.Printf("LISTEN :%s -> %s (retryCount=%d)\n", portStr, target, retryCount)
		retryCount++
		if err := http.ListenAndServe(":"+portStr, forwardRequestHandler); err != nil {
			log.Printf("ERROR on :%s -> %s\n%s\n", portStr, target, err.Error())
		}
	}

	log.Printf("EXITING :%s -> %s\n", portStr, target)
	return nil
}

func newRequestForwarder(port int, targetURL *url.URL, httpClient http.Client) http.HandlerFunc {
	reqCount := 0
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := reqCount
		reqCount++

		r.URL.Host = targetURL.Host
		r.URL.Scheme = targetURL.Scheme
		newURL := r.URL.String()

		log.Printf("[:%d #%d] %s %s\n", port, reqID, r.Method, newURL)

		forwardRequest, err := http.NewRequest(r.Method, newURL, r.Body)
		for headerName, headerValueSlice := range r.Header {
			for _, headerValue := range headerValueSlice {
				forwardRequest.Header.Add(headerName, headerValue)
			}
		}
		if err != nil {
			fmt.Printf("error creating forward request: %v\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		res, err := httpClient.Do(forwardRequest)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		for headerName, headerValueSlice := range res.Header {
			for _, headerValue := range headerValueSlice {
				w.Header().Add(headerName, headerValue)
			}
		}

		log.Printf("[:%d #%d] %s\n", port, reqID, res.Status)

		if res.Body != nil {
			defer res.Body.Close()
			_, err := io.Copy(w, res.Body)
			if err != nil {
				fmt.Printf("error sending response to client: %v\n", err)
			}
		}
	}
}
