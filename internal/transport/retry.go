package transport

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"syscall"
)

func safeToRetry(method, key string) bool {
	return method == http.MethodGet || method == http.MethodHead || key != ""
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout,
		http.StatusTooEarly,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func retryableTransportError(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var urlError *url.Error
	if errors.As(err, &urlError) {
		return retryableTransportError(urlError.Err)
	}
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		if dnsError.IsNotFound {
			return false
		}
		if dnsError.IsTimeout || dnsError.IsTemporary {
			return true
		}
	}
	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return true
	}
	return errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED)
}
