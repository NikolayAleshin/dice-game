package service

import (
	"context"
	"crypto/sha256"
	"dice-game/pkg/domain/model"
	"dice-game/pkg/domain/repository"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GameService struct {
	randomService RandomServiceInterface
	gameRepo      repository.GameRepository
}

func NewGameService(randomService RandomServiceInterface, gameRepo repository.GameRepository) *GameService {
	return &GameService{
		randomService: randomService,
		gameRepo:      gameRepo,
	}
}

func (s *GameService) PlayGame(ctx context.Context, playerID string) (*model.GameResult, error) {
	generator, err := s.randomService.GetRandomGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to get random generator: %w", err)
	}

	playerDice, err := generator.Generate(1, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate player dice: %w", err)
	}

	serverDice, err := generator.Generate(1, 6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server dice: %w", err)
	}

	var winner model.Winner
	if playerDice > serverDice {
		winner = model.WinnerPlayer
	} else if serverDice > playerDice {
		winner = model.WinnerServer
	} else {
		winner = model.WinnerDraw
	}

	now := time.Now()
	gameID := uuid.New().String()

	var verificationKey string
	if verifiableGenerator, ok := generator.(VerifiableGenerator); ok {
		verificationKey = verifiableGenerator.GetVerificationData()
	}

	result := &model.GameResult{
		GameID:          gameID,
		PlayerID:        playerID,
		PlayerDice:      playerDice,
		ServerDice:      serverDice,
		Winner:          winner,
		PlayedAt:        now,
		GeneratorUsed:   generator.Name(),
		VerificationKey: verificationKey,
	}

	if err := s.gameRepo.SaveGameResult(ctx, result); err != nil {
		return nil, fmt.Errorf("failed to save game result: %w", err)
	}

	return result, nil
}

func (s *GameService) GetGameResult(ctx context.Context, gameID string) (*model.GameResult, error) {
	return s.gameRepo.GetGameResult(ctx, gameID)
}

func (s *GameService) VerifyGame(ctx context.Context, gameID, clientSeed string) (bool, error) {
	result, err := s.gameRepo.GetGameResult(ctx, gameID)
	if err != nil {
		return false, fmt.Errorf("failed to get game result: %w", err)
	}

	if result.GeneratorUsed != "provably_fair" {
		return false, fmt.Errorf("game was not played with a verifiable generator")
	}

	if result.VerificationKey == "" {
		return false, fmt.Errorf("verification data is missing for this game")
	}

	parts := strings.Split(result.VerificationKey, ":")
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid verification data format")
	}

	serverSeed := parts[0]
	nonce, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, fmt.Errorf("invalid nonce in verification data: %w", err)
	}
	originalHash := parts[2]

	combinedData := fmt.Sprintf("%s:%s:%d", serverSeed, clientSeed, nonce)
	hasher := sha256.New()
	hasher.Write([]byte(combinedData))
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))

	if calculatedHash != originalHash {
		return false, nil
	}

	playerDice, err := calculateDiceValue(calculatedHash[:8], 1, 6)
	if err != nil {
		return false, fmt.Errorf("failed to calculate player dice: %w", err)
	}

	serverDice, err := calculateDiceValue(calculatedHash[8:16], 1, 6)
	if err != nil {
		return false, fmt.Errorf("failed to calculate server dice: %w", err)
	}

	if playerDice != result.PlayerDice || serverDice != result.ServerDice {
		return false, nil
	}

	return true, nil
}

func calculateDiceValue(hexPart string, min, max int) (int, error) {
	num, err := hex.DecodeString(hexPart)
	if err != nil {
		return 0, fmt.Errorf("failed to decode hex: %w", err)
	}

	var value int = 0
	for _, b := range num {
		value = (value << 8) | int(b)
	}

	rangeSize := max - min + 1
	result := (value % rangeSize) + min

	return result, nil
}

func hexToDiceValue(hexPart string) (int, error) {
	value, err := strconv.ParseInt(hexPart, 16, 64)
	if err != nil {
		return 0, err
	}

	return int(value%6) + 1, nil
}
