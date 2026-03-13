package telegram

import (
	"context"
	"errors"
	"io"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type ImageService struct {
	client *telegram.Client
	logger *zap.Logger

	downloader *downloader.Downloader
}

func NewImageService(client *telegram.Client, logger *zap.Logger) *ImageService {
	return &ImageService{
		client:     client,
		logger:     logger,
		downloader: downloader.NewDownloader(),
	}
}

func (i *ImageService) GetChannelsMessageImageById(ctx context.Context, channelId string, messageId int) (io.Reader, error) {
	api := i.client.API()

	resolvedChats, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: channelId,
	})
	if err != nil {
		return nil, err
	}

	if len(resolvedChats.Chats) == 0 {
		return nil, errors.New("channel not found")
	}

	channel, ok := resolvedChats.Chats[0].(*tg.Channel)
	if !ok {
		return nil, errors.New("resolved peer is not a channel")
	}

	messages, err := api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
		Channel: channel.AsInput(),
		ID: []tg.InputMessageClass{
			&tg.InputMessageID{ID: messageId},
		},
	})

	if err != nil {
		return nil, errors.New("unable to fetch messages from channel")
	}

	msgs, ok := messages.(*tg.MessagesChannelMessages)
	if !ok {
		return nil, errors.New("messages fetched do not represent a channel message")
	}

	// Count will always be 1, because we are giving one ID and even if its not found, an empty message is returned
	message, ok := msgs.Messages[0].(*tg.Message)
	if !ok {
		return nil, errors.New("no message found")
	}

	media, ok := message.Media.(*tg.MessageMediaPhoto)
	if !ok {
		return nil, errors.New("the message does not contain a photo media object")
	}

	photo, ok := media.Photo.(*tg.Photo)
	if !ok {
		return nil, errors.New("the message media has no photo payload")
	}

	var largestSize *tg.PhotoSize
	for _, size := range photo.Sizes {
		if ps, ok := size.(*tg.PhotoSize); ok {
			if largestSize == nil || ps.Size > largestSize.Size {
				largestSize = size.(*tg.PhotoSize)
			}
		}
	}
	if largestSize == nil {
		return nil, errors.New("no valid photo size found")
	}
	photoLocation := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,

		ThumbSize: largestSize.Type,
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, err := i.downloader.Download(api, photoLocation).Stream(ctx, pw)
		if err != nil {
			pw.CloseWithError(errors.New("unable to download photo"))
		}
	}()

	return pr, nil
}
