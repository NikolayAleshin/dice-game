package usecase

import (
	"context"
	"dice-game/pkg/domain/model"
	"dice-game/pkg/domain/service"
)

type GameUseCase struct {
	gameService service.GameServiceInterface
}

func NewGameUseCase(gameService service.GameServiceInterface) *GameUseCase {
	return &GameUseCase{
		gameService: gameService,
	}
}

func (uc *GameUseCase) PlayGame(ctx context.Context, playerID string) (*model.GameResult, error) {
	if playerID == "" {
		playerID = "anonymous"
	}

	return uc.gameService.PlayGame(ctx, playerID)
}

func (uc *GameUseCase) VerifyGame(ctx context.Context, gameID, verificationData string) (bool, error) {
	return uc.gameService.VerifyGame(ctx, gameID, verificationData)
}

func (uc *GameUseCase) GetGameResult(ctx context.Context, gameID string) (*model.GameResult, error) {
	return uc.gameService.GetGameResult(ctx, gameID)
}
