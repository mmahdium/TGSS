package rss

import (
	"context"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type FeedCache struct {
	logger     *zap.Logger
	feedsCache map[string]*RSSFeed
	mu         sync.Mutex
}

func NewFeedCache(logger *zap.Logger) *FeedCache {
	return &FeedCache{logger: logger, feedsCache: map[string]*RSSFeed{}, mu: sync.Mutex{}}
}

func (c *FeedCache) SetFeedToFeedCache(channelId string, feed *RSSFeed) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.feedsCache[channelId] = feed
}

func (c *FeedCache) GetFeedFromFeedCache(channelId string, limit int) *RSSFeed {
	c.mu.Lock()
	defer c.mu.Unlock()

	if feed, ok := c.feedsCache[channelId]; ok {
		feedBuildTime, err := time.Parse("Mon, 02 Jan 2006 15:04 -0700", feed.Channel.LastBuildDate)
		if err != nil || time.Since(feedBuildTime) > time.Minute*15 {
			delete(c.feedsCache, channelId)
			return nil
		}

		if len(feed.Channel.Items) >= limit {
			feedCopy := *feed
			feedCopy.Channel.Items = feed.Channel.Items[len(feed.Channel.Items)-limit:]
			c.logger.Debug("Cache hit on channel ID", zap.String("channelId", channelId))
			return &feedCopy
		}
		return nil
	}
	return nil
}

func (c *FeedCache) CleanupFeedCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for channelId, feed := range c.feedsCache {
		feedBuildTime, _ := time.Parse("Mon, 02 Jan 2006 15:04 MST", feed.Channel.LastBuildDate)
		if time.Since(feedBuildTime) > time.Minute*15 {
			delete(c.feedsCache, channelId)
		}

	}

}

func RegisterFeedCacheCleanup(lc fx.Lifecycle, logger *zap.Logger, f *FeedCache) {
	var ticker *time.Ticker
	var stopChan chan struct{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting feed cache cleanup task", zap.Duration("interval", 30*time.Minute))
			ticker = time.NewTicker(20 * time.Minute)
			stopChan = make(chan struct{})
			go func() {
				for {
					select {
					case <-ticker.C:
						logger.Info("Running feed cache cleanup")
						f.CleanupFeedCache()
						logger.Info("Feed cache cleanup completed")
					case <-stopChan:
						return
					}
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping feed cache cleanup task")
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
