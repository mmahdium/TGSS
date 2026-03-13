package server

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// TODO: another limiter rate for images
type ClientLimiter struct {
	feedLimiter  *rate.Limiter
	imageLimiter *rate.Limiter
	lastSeen     time.Time
}

type RateLimiter struct {
	logger  *zap.Logger
	clients map[string]*ClientLimiter
	mu      sync.Mutex
}

func NewRateLimiter(logger *zap.Logger) *RateLimiter {
	return &RateLimiter{logger: logger, mu: sync.Mutex{}, clients: map[string]*ClientLimiter{}}
}

func (r *RateLimiter) getLimiter(ip string, kind string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c, ok := r.clients[ip]; ok {
		c.lastSeen = time.Now()
		if kind == "image" {
			return c.imageLimiter
		}
		return c.feedLimiter
	}

	feedLim := rate.NewLimiter(rate.Every(3*time.Second), 5)
	imageLim := rate.NewLimiter(rate.Every(5*time.Second), 4)

	r.clients[ip] = &ClientLimiter{
		feedLimiter:  feedLim,
		imageLimiter: imageLim,
		lastSeen:     time.Now(),
	}
	if kind == "image" {
		return imageLim
	}
	return feedLim
}

func (r *RateLimiter) CleanupRateLimiter() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for ip, clientLimiter := range r.clients {
		if time.Since(clientLimiter.lastSeen) > time.Hour { // Delete ip ratelimiters older that 1 hour
			delete(r.clients, ip)
		}

	}

}

func RegisterRateLimiterCleanup(lc fx.Lifecycle, logger *zap.Logger, r *RateLimiter) {
	var ticker *time.Ticker
	var stopChan chan struct{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting rate limiter cleanup task", zap.Duration("interval", 30*time.Minute))
			ticker = time.NewTicker(30 * time.Minute)
			stopChan = make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						logger.Info("Running rate limiter cleanup")
						r.CleanupRateLimiter()
						logger.Info("Rate limiter cleanup completed")
					case <-stopChan:
						return
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping rate limiter cleanup task")
			if ticker != nil {
				ticker.Stop()
			}
			if stopChan != nil {
				close(stopChan)
			}
			return nil
		},
	})
}

func (r *RateLimiter) FeedRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		lim := r.getLimiter(ip, "feed")
		if !lim.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "Slow down pls"})
			return
		}
		c.Next()
	}
}

func (r *RateLimiter) ImageRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		lim := r.getLimiter(ip, "image")
		if !lim.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "Slow down pls, these arent your mothers nudes"})
			return
		}
		c.Next()
	}
}
