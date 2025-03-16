-- Create game_results table
CREATE TABLE IF NOT EXISTS game_results (
                                            id SERIAL PRIMARY KEY,
                                            game_id VARCHAR(36) NOT NULL UNIQUE,
    player_id VARCHAR(100) NOT NULL,
    player_dice INTEGER NOT NULL CHECK (player_dice >= 1 AND player_dice <= 6),
    server_dice INTEGER NOT NULL CHECK (server_dice >= 1 AND server_dice <= 6),
    winner VARCHAR(10) NOT NULL CHECK (winner IN ('PLAYER', 'SERVER', 'DRAW')),
    played_at TIMESTAMP WITH TIME ZONE NOT NULL,
                            generator_used VARCHAR(50) NOT NULL,
    verification_key TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                            );

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_game_results_player_id ON game_results(player_id);
CREATE INDEX IF NOT EXISTS idx_game_results_played_at ON game_results(played_at);