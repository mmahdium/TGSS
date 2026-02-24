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
	if len(resolved.Chats) == 0 {
		return nil, errors.New("channel not found")
	}
	channel, ok := resolved.Chats[0].(*tg.Channel)
	if !ok {
		return nil, errors.New("resolved peer is not a channel")
	}

	history, err := s.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channel.ID,
			AccessHash: channel.AccessHash,
		},
		Limit: 100,
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
