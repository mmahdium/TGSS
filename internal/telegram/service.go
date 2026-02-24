package telegram

import (
	"context"
	"errors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Service struct {
	client *telegram.Client
	log    *zap.Logger
}

func (s *Service) LastMessages(ctx context.Context, username string, limit int) ([]tg.MessageClass, error) {
	api := s.client.API()
	resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	var peer tg.InputPeerClass
	for _, chat := range resolved.Chats {
		if ch, ok := chat.(*tg.Channel); ok {
			peer = &tg.InputPeerChannel{
				ChannelID:  ch.ID,
				AccessHash: ch.AccessHash,
			}
			break
		}
	}
	if peer == nil {
		return nil, errors.New("channel not found")
	}
	// Fetch history
	history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  peer,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	// Extract messages
	var msgs []tg.MessageClass
	switch h := history.(type) {
	case *tg.MessagesMessages:
		msgs = h.Messages
	case *tg.MessagesMessagesSlice:
		msgs = h.Messages
	case *tg.MessagesChannelMessages:
		msgs = h.Messages
	}
	return msgs, nil
}
