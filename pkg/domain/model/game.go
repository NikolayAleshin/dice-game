package model

import "time"

type Winner string

const (
	WinnerPlayer Winner = "PLAYER"
	WinnerServer Winner = "SERVER"
	WinnerDraw   Winner = "DRAW"
)

type GameResult struct {
	GameID          string
	PlayerID        string
	PlayerDice      int
	ServerDice      int
	Winner          Winner
	PlayedAt        time.Time
	GeneratorUsed   string
	VerificationKey string
}
