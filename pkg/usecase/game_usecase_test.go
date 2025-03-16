package usecase

import (
	"context"
	"dice-game/pkg/domain/model"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) PlayGame(ctx context.Context, playerID string) (*model.GameResult, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GameResult), args.Error(1)
}

func (m *MockGameService) GetGameResult(ctx context.Context, gameID string) (*model.GameResult, error) {
	args := m.Called(ctx, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GameResult), args.Error(1)
}

func (m *MockGameService) VerifyGame(ctx context.Context, gameID, clientSeed string) (bool, error) {
	args := m.Called(ctx, gameID, clientSeed)
	return args.Bool(0), args.Error(1)
}

func TestNewGameUseCase(t *testing.T) {
	// Arrange
	mockService := new(MockGameService)

	// Act
	usecase := NewGameUseCase(mockService)

	// Assert
	assert.NotNil(t, usecase)
	assert.Equal(t, mockService, usecase.gameService)
}

func TestGameUseCase_PlayGame(t *testing.T) {
	t.Run("Success with valid player ID", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		expectedResult := &model.GameResult{
			GameID:     "test-game-id",
			PlayerID:   "test-player",
			PlayerDice: 4,
			ServerDice: 2,
			Winner:     model.WinnerPlayer,
			PlayedAt:   time.Now(),
		}

		mockService.On("PlayGame", mock.Anything, "test-player").Return(expectedResult, nil)
		usecase := NewGameUseCase(mockService)

		// Act
		result, err := usecase.PlayGame(context.Background(), "test-player")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockService.AssertExpectations(t)
	})

	t.Run("Success with empty player ID (anonymous)", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		expectedResult := &model.GameResult{
			GameID:     "test-game-id",
			PlayerID:   "anonymous",
			PlayerDice: 3,
			ServerDice: 5,
			Winner:     model.WinnerServer,
			PlayedAt:   time.Now(),
		}

		mockService.On("PlayGame", mock.Anything, "anonymous").Return(expectedResult, nil)
		usecase := NewGameUseCase(mockService)

		// Act
		result, err := usecase.PlayGame(context.Background(), "")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockService.AssertExpectations(t)
	})

	t.Run("Error from service", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		expectedError := errors.New("service error")

		mockService.On("PlayGame", mock.Anything, "test-player").Return(nil, expectedError)
		usecase := NewGameUseCase(mockService)

		// Act
		result, err := usecase.PlayGame(context.Background(), "test-player")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockService.AssertExpectations(t)
	})
}

func TestGameUseCase_GetGameResult(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		expectedResult := &model.GameResult{
			GameID:     "test-game-id",
			PlayerID:   "test-player",
			PlayerDice: 4,
			ServerDice: 2,
			Winner:     model.WinnerPlayer,
			PlayedAt:   time.Now(),
		}

		mockService.On("GetGameResult", mock.Anything, "test-game-id").Return(expectedResult, nil)
		usecase := NewGameUseCase(mockService)

		// Act
		result, err := usecase.GetGameResult(context.Background(), "test-game-id")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockService.AssertExpectations(t)
	})

	t.Run("Game not found", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		expectedError := errors.New("game not found")

		mockService.On("GetGameResult", mock.Anything, "nonexistent-id").Return(nil, expectedError)
		usecase := NewGameUseCase(mockService)

		// Act
		result, err := usecase.GetGameResult(context.Background(), "nonexistent-id")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockService.AssertExpectations(t)
	})
}

func TestGameUseCase_VerifyGame(t *testing.T) {
	t.Run("Successful verification", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		gameID := "test-game-id"
		clientSeed := "test-client-seed"

		mockService.On("VerifyGame", mock.Anything, gameID, clientSeed).Return(true, nil)
		usecase := NewGameUseCase(mockService)

		// Act
		isValid, err := usecase.VerifyGame(context.Background(), gameID, clientSeed)

		// Assert
		assert.NoError(t, err)
		assert.True(t, isValid)
		mockService.AssertExpectations(t)
	})

	t.Run("Failed verification", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		gameID := "test-game-id"
		clientSeed := "test-client-seed"

		mockService.On("VerifyGame", mock.Anything, gameID, clientSeed).Return(false, nil)
		usecase := NewGameUseCase(mockService)

		// Act
		isValid, err := usecase.VerifyGame(context.Background(), gameID, clientSeed)

		// Assert
		assert.NoError(t, err)
		assert.False(t, isValid)
		mockService.AssertExpectations(t)
	})

	t.Run("Verification error", func(t *testing.T) {
		// Arrange
		mockService := new(MockGameService)
		gameID := "test-game-id"
		clientSeed := "test-client-seed"
		expectedError := errors.New("verification error")

		mockService.On("VerifyGame", mock.Anything, gameID, clientSeed).Return(false, expectedError)
		usecase := NewGameUseCase(mockService)

		// Act
		isValid, err := usecase.VerifyGame(context.Background(), gameID, clientSeed)

		// Assert
		assert.Error(t, err)
		assert.False(t, isValid)
		assert.Equal(t, expectedError, err)
		mockService.AssertExpectations(t)
	})
}

func TestGameUseCase_ContextPropagation(t *testing.T) {
	type ctxKey string
	var testKey ctxKey = "test-key"
	testValue := "test-value"
	ctx := context.WithValue(context.Background(), testKey, testValue)

	mockService := new(MockGameService)
	mockService.On("PlayGame", mock.MatchedBy(func(c context.Context) bool {
		return c.Value(testKey) == testValue
	}), "test-player").Return(&model.GameResult{}, nil)

	usecase := NewGameUseCase(mockService)

	// Act
	_, err := usecase.PlayGame(ctx, "test-player")

	// Assert
	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
