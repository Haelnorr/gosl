package slapshotapi

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"
)

type slapAPIClient struct {
	client      *http.Client
	ratelimiter *rate.Limiter
	mu          sync.Mutex
	maxTokens   int
}

func (c *slapAPIClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	for {
		err := c.ratelimiter.Wait(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "c.ratelimiter.Wait")
		}
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "c.client.Do")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resetAfter := 30 * time.Second
			resp.Body.Close()
			if resetAfter > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(resetAfter):
					continue
				}
			}
		}
		c.updateLimiterFromHeaders(resp.Header)
		return resp, nil
	}
}

func (c *slapAPIClient) updateLimiterFromHeaders(h http.Header) {
	c.mu.Lock()
	defer c.mu.Unlock()

	limit, err1 := strconv.Atoi(h.Get("RateLimit-Limit"))
	window, err2 := strconv.Atoi(h.Get("RateLimit-Window"))

	if err1 != nil || err2 != nil || limit <= 0 || window <= 0 {
		return
	}

	if limit != c.maxTokens || time.Duration(window) != time.Duration(float64(window)/float64(limit))*time.Second {
		c.maxTokens = limit
		c.ratelimiter.SetBurst(limit)
		c.ratelimiter.SetLimit(rate.Every(time.Duration(window) / time.Duration(limit)))
	}
}

func newRateLimitedClient() *slapAPIClient {
	rl := rate.NewLimiter(rate.Inf, 10)
	c := &slapAPIClient{
		client:      http.DefaultClient,
		ratelimiter: rl,
		maxTokens:   10,
	}
	return c
}
