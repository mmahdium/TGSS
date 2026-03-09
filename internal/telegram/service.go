package telegram

import (
	"context"
	"errors"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Service struct {
	client *telegram.Client
	log    *zap.Logger

	mu            sync.Mutex
	phone         string
	phoneCodeHash string
}

func NewService(client *telegram.Client, logger *zap.Logger) *Service {
	return &Service{
		client: client,
		log:    logger,
	}
}

func (s *Service) AuthStatus(ctx context.Context) (bool, error) {
	_, err := s.client.Auth().Status(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) SendCode(ctx context.Context, phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.client.Auth().SendCode(ctx, phone, auth.SendCodeOptions{})
	if err != nil {
		return err
	}

	s.phone = phone
	s.phoneCodeHash = resp.(*tg.AuthSentCode).PhoneCodeHash

	return nil
}

func (s *Service) VerifyCode(ctx context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.phone == "" {
		return errors.New("phone not initialized")
	}

	_, err := s.client.Auth().SignIn(ctx, s.phone, code, s.phoneCodeHash)
	return err
}

func (s *Service) Password(ctx context.Context, password string) error {
	_, err := s.client.Auth().Password(ctx, password)
	return err
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

	history, err := api.MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChannel{
			ChannelID:  channel.ID,
			AccessHash: channel.AccessHash,
		},
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

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
