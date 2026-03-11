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

type ClientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	logger  *zap.Logger
	clients map[string]*ClientLimiter
	mu      sync.Mutex
}

func NewRateLimiter(logger *zap.Logger) *RateLimiter {
	return &RateLimiter{logger: logger, mu: sync.Mutex{}, clients: map[string]*ClientLimiter{}}
}

func (r *RateLimiter) getLimiter(ip string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c, ok := r.clients[ip]; ok {
		c.lastSeen = time.Now()
		return c.limiter
	}

	limiter := rate.NewLimiter(rate.Every(time.Second*3), 5) // TODO: get from config
	r.clients[ip] = &ClientLimiter{limiter: limiter, lastSeen: time.Now()}

	return limiter
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

func RegisterCleanup(lc fx.Lifecycle, logger *zap.Logger, r *RateLimiter) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting rate limiter cleanup task", zap.Duration("interval", 30*time.Minute))
			ticker := time.NewTicker(30 * time.Minute)
			go func() {
				for range ticker.C {
					logger.Info("Running rate limiter cleanup")
					r.CleanupRateLimiter()
					logger.Info("Rate limiter cleanup completed")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			r.logger.Info("Stopping rate limiter cleanup task")
			return nil
		},
	})
}

func (r *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		limiter := r.getLimiter(ip)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "Slow down pls"})
			return
		}
		c.Next()
	}
}
