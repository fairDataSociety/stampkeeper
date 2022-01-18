package api

import (
	"context"

	"github.com/fairDataSociety/stampkeeper/pkg/bot"

	"github.com/fairDataSociety/stampkeeper/pkg/keeper"
	"github.com/fairDataSociety/stampkeeper/pkg/logging"
)

type Handler struct {
	stampkeeper *keeper.Keeper
	logger      logging.Logger

	bot bot.Bot
}

func NewHandler(ctx context.Context, serverEndpoint string, logger logging.Logger) *Handler {
	return &Handler{
		stampkeeper: keeper.New(ctx, serverEndpoint, logger),
		logger:      logger,
	}
}

func (h *Handler) SetBot(b bot.Bot) {
	h.bot = b
}
