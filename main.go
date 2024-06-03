package ratelimiter

import (
	"github.com/gin-gonic/gin"

	"time"
	"sync"
	"net/http"
)

type RateLimiter struct {
	// Should return an unique key for each client
	// Default: gin.Context.ClientIP()
	Key func(*gin.Context) string

	// Default: gin.Context.AbortWithStatus(http.StatusTooManyRequests)
	WhenLimitReached func(*gin.Context, *Client)

	// Maximum number of requests allowed within the rate window
	// Default: 50
	Limit int

	// Default: time.Second
	Rate time.Duration

	store map[string]*Client
	mutex *sync.Mutex
}

type Options struct {
	Key func(*gin.Context) string
	WhenLimitReached func(*gin.Context, *Client)
	Limit int
	Rate time.Duration
}

type Client struct {
	Key string
	Visits int
	IsBanned bool
	BannedUntil time.Time

	rateLimiter *RateLimiter
}

func defaultKey(c *gin.Context) string {
	return c.ClientIP()
}

func defaultWhenLimitReached(c *gin.Context, client *Client) {
	c.AbortWithStatus(http.StatusTooManyRequests)
}

func Default() *RateLimiter {
	return New(Options{})
}

func New(o Options) *RateLimiter {
	r := &RateLimiter{
		Key: o.Key,
		WhenLimitReached: o.WhenLimitReached,
		Limit: o.Limit,
		Rate: o.Rate,
		store: make(map[string]*Client),
		mutex: &sync.Mutex{},
	}

	if r.Key == nil {
		r.Key = defaultKey
	}

	if r.WhenLimitReached == nil {
		r.WhenLimitReached = defaultWhenLimitReached
	}

	if r.Limit == 0 {
		r.Limit = 50
	}

	if r.Rate == 0 {
		r.Rate = time.Second
	}

	return r
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := r.Key(c)

		r.mutex.Lock()
		defer r.mutex.Unlock()

		if _, ok := r.store[key]; !ok {
			r.store[key] = &Client{
				Key: key,
				Visits: 0,
				rateLimiter: r,
			}
		}

		client := r.store[key]

		if client.Visits >= r.Limit || client.IsBanned {
			r.WhenLimitReached(c, client)
			return
		}

		client.Visits +=  1

		time.AfterFunc(r.Rate, func() {
			r.mutex.Lock()
			defer r.mutex.Unlock()

			client.Visits -= 1
		})

		c.Next()
	}
}

func (r *RateLimiter) Client(key string) *Client {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.store[key]
}

func (c *Client) Ban(duration time.Duration) {
	c.rateLimiter.mutex.Lock()
	defer c.rateLimiter.mutex.Unlock()

	c.IsBanned = true
	c.BannedUntil = time.Now().Add(duration)

	time.AfterFunc(duration, func() {
		c.rateLimiter.mutex.Lock()
		defer c.rateLimiter.mutex.Unlock()

		c.Unban()
	})
}

func (c *Client) Unban() {
	c.rateLimiter.mutex.Lock()
	defer c.rateLimiter.mutex.Unlock()

	c.IsBanned = false
	c.BannedUntil = time.Time{}
}