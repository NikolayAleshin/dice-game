package service

import (
	"context"
	"dice-game/pkg/domain/model"
)

type GameServiceInterface interface {
	PlayGame(ctx context.Context, playerID string) (*model.GameResult, error)
	VerifyGame(ctx context.Context, gameID string, verificationData string) (bool, error)
	GetGameResult(ctx context.Context, gameID string) (*model.GameResult, error)
}
