package rss

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate,omitempty"`
	Description string `xml:"description"`
	Guid        string `xml:"guid,omitempty"`
}

type RSSChannel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	PubDate       string `xml:"pubDate,omitempty"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`
	Generator     string `xml:"generator,omitempty"`

	Items []RSSItem `xml:"item"`
}

type RSSFeed struct {
	XMLName   xml.Name `xml:"rss"`
	Version   string   `xml:"version,attr"`
	XmlnsAtom string   `xml:"xmlns:atom,attr"`

	Channel RSSChannel `xml:"channel"`
}

type RSSGenerator struct {
	logger *zap.Logger
}

func NewRSSGenerator(logger *zap.Logger) *RSSGenerator {
	return &RSSGenerator{logger: logger}
}

func (r *RSSGenerator) GenerateFeed(items []tg.MessageClass, channelId string) *RSSFeed {
	if channelId == "" {
		r.logger.Error("GenerateFeed called with empty channelId")
		return nil
	}

	nowStr := time.Now().Format("Mon, 02 Jan 2006 15:04 MST")

	rssChannel := &RSSChannel{
		Title: "Recent posts from @" + channelId,
		Link:  "https://t.me/" + channelId,
		Description: "This feed contains the most recent posts from the Telegram channel @" + channelId + ". " +
			"Stay updated with the latest news and updates from the channel.",
		PubDate:       nowStr,
		LastBuildDate: nowStr,
		Generator:     "Telegram RSS Generator",
	}

	errorCount := 0

	for _, m := range items {
		if m == nil {
			errorCount++
			continue
		}

		item, err := r.messageToItem(m, channelId)
		if err != nil {
			r.logger.Error("failed to convert message to RSS item",
				zap.String("channel", channelId),
				zap.Int("message_id", m.GetID()),
				zap.Error(err),
			)
			errorCount++
			continue
		}

		rssChannel.Items = append(rssChannel.Items, *item)
	}

	if errorCount > 0 {
		r.logger.Warn("feed generation completed with errors",
			zap.String("channel", channelId),
			zap.Int("failed", errorCount),
		)
	}

	return &RSSFeed{
		Version:   "2.0",
		XmlnsAtom: "http://www.w3.org/2005/Atom",
		Channel:   *rssChannel,
	}
}

func (r *RSSGenerator) messageToItem(msg tg.MessageClass, channelId string) (*RSSItem, error) {
	message, ok := msg.(*tg.Message)
	if !ok {
		return nil, fmt.Errorf("unsupported message type: %T", msg)
	}

	messageURL, err := url.Parse("https://t.me/" + channelId + "/" + strconv.Itoa(msg.GetID()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse message URL: %w", err)
	}

	description := message.Message
	if description == "" {
		description = "No content"
	}

	return &RSSItem{
		Title:       "Post by @" + channelId + " on Telegram",
		Link:        messageURL.String(),
		PubDate:     time.Unix(int64(message.Date), 0).Format("Mon, 02 Jan 2006 15:04 MST"),
		Description: description,
		Guid:        messageURL.String(),
	}, nil
}
