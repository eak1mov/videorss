package main

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Throttler struct {
	roundTripper http.RoundTripper
	rateLimiter  *rate.Limiter
}

func (t *Throttler) RoundTrip(r *http.Request) (*http.Response, error) {
	err := t.rateLimiter.Wait(r.Context())
	if err != nil {
		return nil, err
	}
	return t.roundTripper.RoundTrip(r)
}

func NewThrottler(limit int, period time.Duration, roundTripper http.RoundTripper) http.RoundTripper {
	return &Throttler{
		roundTripper: roundTripper,
		rateLimiter:  rate.NewLimiter(rate.Every(period), limit),
	}
}
