package telegram

import (
	"tgss/internal/config"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"go.uber.org/fx"
	"go.uber.org/zap"
)  
  
func NewTelegramClient(cfg *config.Config, logger *zap.Logger) *telegram.Client {  
    return telegram.NewClient(cfg.TgAppId, cfg.TgAppHash, telegram.Options{  
        DC:     2,  
        DCList: dcs.Prod(),  
        Logger: logger,  
        SessionStorage: &session.FileStorage{  
            Path: cfg.SessionPath,  
        },  
    })  
}  
  
func NewService(client *telegram.Client, logger *zap.Logger) *Service {  
    return &Service{client: client, log: logger}  
}  
  
var Module = fx.Provide(  
    NewTelegramClient,  
    NewService,  
)