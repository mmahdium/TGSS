package rss

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"tgss/internal/config"
	"time"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type RSSItem struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Enclosure   *RSSEnclosure `xml:"enclosure,omitempty"`
	PubDate     string        `xml:"pubDate,omitempty"`
	Description string        `xml:"description"`
	Guid        string        `xml:"guid,omitempty"`
}

type RSSEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"lenghth,omitempty,attr"`
	Type   string `xml:"type,attr"`
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
	config *config.Config
}

func NewRSSGenerator(logger *zap.Logger, config *config.Config) *RSSGenerator {
	return &RSSGenerator{logger: logger, config: config}
}

func (r *RSSGenerator) GenerateFeed(items []tg.MessageClass, channelId string) *RSSFeed {
	if channelId == "" {
		r.logger.Error("GenerateFeed called with empty channelId")
		return nil
	}

	nowStr := time.Now().Format("Mon, 02 Jan 2006 15:04 MST")

	rssChannel := &RSSChannel{
		Title:         "Recent posts from @" + channelId,
		Link:          "https://t.me/" + channelId,
		Description:   "Recent posts from the Telegram channel @" + channelId + ".",
		PubDate:       nowStr,
		LastBuildDate: nowStr,
		Generator:     "TGSS",
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

	// Logging is done by AI
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

	rssItem := &RSSItem{
		Title:       "Post by @" + channelId + " on Telegram",
		Link:        messageURL.String(),
		PubDate:     time.Unix(int64(message.Date), 0).Format("Mon, 02 Jan 2006 15:04 MST"),
		Description: description,
		Guid:        messageURL.String(),
	}

	if err = r.messageHasPhoto(message); err == nil {
		if enclosureURL, err := url.ParseRequestURI(r.config.BaseURL + "/image/" + channelId + "/" + strconv.Itoa(msg.GetID())); err == nil {
			rssItem.Enclosure = &RSSEnclosure{
				URL:    enclosureURL.String(),
				Length: "0",
				Type:   "image/jpeg",
			}
		}
	}

	return rssItem, nil
}

func (r *RSSGenerator) messageHasPhoto(message tg.MessageClass) error {
	media, ok := message.(*tg.Message).Media.(*tg.MessageMediaPhoto)
	if !ok {
		return errors.New("the message does not contain a photo media object")
	}

	_, ok = media.Photo.(*tg.Photo)
	if !ok {
		return errors.New("the message media has no photo payload")
	}

	return nil
}
