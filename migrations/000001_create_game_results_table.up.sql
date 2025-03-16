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

CREATE INDEX IF NOT EXISTS idx_game_results_player_id ON game_results(player_id);
CREATE INDEX IF NOT EXISTS idx_game_results_played_at ON game_results(played_at);

CREATE TABLE IF NOT EXISTS game_statistics (
                                               id SERIAL PRIMARY KEY,
                                               date DATE NOT NULL UNIQUE,
                                               total_games INTEGER NOT NULL DEFAULT 0,
                                               player_wins INTEGER NOT NULL DEFAULT 0,
                                               server_wins INTEGER NOT NULL DEFAULT 0,
                                               draws INTEGER NOT NULL DEFAULT 0,
                                               updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS verification_records (
                                                    id SERIAL PRIMARY KEY,
                                                    game_id VARCHAR(36) NOT NULL REFERENCES game_results(game_id),
    verification_data TEXT NOT NULL,
    is_valid BOOLEAN NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
                              );

CREATE OR REPLACE FUNCTION update_game_statistics()
RETURNS TRIGGER AS $$
DECLARE
game_date DATE;
BEGIN
    game_date := DATE(NEW.played_at);

UPDATE game_statistics
SET
    total_games = total_games + 1,
    player_wins = CASE WHEN NEW.winner = 'PLAYER' THEN player_wins + 1 ELSE player_wins END,
    server_wins = CASE WHEN NEW.winner = 'SERVER' THEN server_wins + 1 ELSE server_wins END,
    draws = CASE WHEN NEW.winner = 'DRAW' THEN draws + 1 ELSE draws END,
    updated_at = CURRENT_TIMESTAMP
WHERE date = game_date;

IF NOT FOUND THEN
        INSERT INTO game_statistics (
            date,
            total_games,
            player_wins,
            server_wins,
            draws
        ) VALUES (
            game_date,
            1,
            CASE WHEN NEW.winner = 'PLAYER' THEN 1 ELSE 0 END,
            CASE WHEN NEW.winner = 'SERVER' THEN 1 ELSE 0 END,
            CASE WHEN NEW.winner = 'DRAW' THEN 1 ELSE 0 END
        );
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_game_statistics
    AFTER INSERT ON game_results
    FOR EACH ROW
    EXECUTE FUNCTION update_game_statistics();

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgresql;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO postgresql;