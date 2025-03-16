package grpc

import (
	"context"
	"dice-game/pkg/domain/interfaces"
	"dice-game/pkg/usecase"
	pb "dice-game/proto/gen"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiceGameService struct {
	pb.UnimplementedDiceGameServiceServer
	gameUseCase usecase.GameUseCaseInterface
	logger      interfaces.Logger
}

func NewDiceGameService(gameUseCase usecase.GameUseCaseInterface, logger interfaces.Logger) *DiceGameService {
	return &DiceGameService{
		gameUseCase: gameUseCase,
		logger:      logger.With().Str("component", "dice_game_grpc_service").Logger(),
	}
}

func (s *DiceGameService) Play(ctx context.Context, req *pb.PlayRequest) (*pb.PlayResponse, error) {
	s.logger.Info().Msg("Received Play request")

	if ctx.Err() != nil {
		return nil, status.Error(codes.DeadlineExceeded, "client context already done")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := s.gameUseCase.PlayGame(ctx, req.GetPlayerId())
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to process play request")
		return nil, status.Errorf(codes.Internal, "failed to process play request: %v", err)
	}

	response := &pb.PlayResponse{
		GameId:          result.GameID,
		PlayerDice:      int32(result.PlayerDice),
		ServerDice:      int32(result.ServerDice),
		Winner:          string(result.Winner),
		PlayedAt:        result.PlayedAt.Format(time.RFC3339),
		GeneratorUsed:   result.GeneratorUsed,
		VerificationKey: result.VerificationKey,
	}

	s.logger.Info().
		Int("player_dice", result.PlayerDice).
		Int("server_dice", result.ServerDice).
		Str("winner", string(result.Winner)).
		Str("game_id", result.GameID).
		Msg("Game completed successfully")

	return response, nil
}

func (s *DiceGameService) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	s.logger.Info().Str("game_id", req.GetGameId()).Msg("Received Verify request")

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	isValid, err := s.gameUseCase.VerifyGame(ctx, req.GetGameId(), req.GetVerificationData())
	if err != nil {
		s.logger.Error().Err(err).Str("game_id", req.GetGameId()).Msg("Failed to verify game")
		return nil, status.Errorf(codes.Internal, "failed to verify game: %v", err)
	}

	response := &pb.VerifyResponse{
		GameId:  req.GetGameId(),
		IsValid: isValid,
	}

	s.logger.Info().
		Str("game_id", req.GetGameId()).
		Bool("is_valid", isValid).
		Msg("Game verification completed")

	return response, nil
}
