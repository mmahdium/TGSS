package rss

import (
	"encoding/xml"
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

func NewRSSGenerator(logger zap.Logger) *RSSGenerator {
	return &RSSGenerator{logger: &logger}
}

func (r *RSSGenerator) GenerateFeed(items []tg.MessageClass, channelId string) *RSSFeed {
	rssChannel := &RSSChannel{
		Title: "Recent posts from @" + channelId,
		Link:  "https://t.me/" + channelId,
		Description: "This feed contains the most recent posts from the Telegram channel @" + channelId + ". " +
			"Stay updated with the latest news and updates from the channel.",
		PubDate:       time.Now().Format("Mon, 02 Jan 2006 15:04 MST"),
		LastBuildDate: time.Now().Format("Mon, 02 Jan 2006 15:04 MST"),
		Generator:     "Telegram RSS Generator",
	}
	for _, m := range items {
		rssChannel.Items = append(rssChannel.Items, *r.messageToItem(m, channelId))
	}
	return &RSSFeed{
		Version:   "2.0",
		XmlnsAtom: "http://www.w3.org/2005/Atom",
		Channel:   *rssChannel,
	}
}

func (r *RSSGenerator) messageToItem(msg tg.MessageClass, channelId string) *RSSItem {
	return &RSSItem{
		Title:       "Post by @" + channelId + " on Telegram",
		Link:        "https://t.me/" + channelId + "/" + strconv.Itoa(msg.GetID()), // TODO: URL parse it
		PubDate:     time.Unix(int64(msg.(*tg.Message).Date), 0).Format("Mon, 02 Jan 2006 15:04 MST"),
		Description: msg.(*tg.Message).Message,
		Guid:        "https://t.me/" + channelId + "/" + strconv.Itoa(msg.GetID()),
	}
}
