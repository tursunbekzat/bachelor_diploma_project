package db

import (
	"backend/data"
	"database/sql"
	"fmt"
	"log"
)

// CreateGame adds a new game to the database
func CreateGame(game *data.Game) error {
    query := `INSERT INTO games (game_name, creator_id, status) VALUES ($1, $2, $3) RETURNING id`
    err := DB.QueryRow(query, game.GameName, game.CreatorID, game.Status).Scan(&game.ID)
    if err != nil {
        return fmt.Errorf("could not insert game: %v", err)
    }
    return nil
}

// Gets list of games with detailed information
func GetAllGames() ([]*data.Game, error) {
    rows, err := DB.Query("SELECT id, game_name, creator_id, status, created_at FROM games")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var games []*data.Game
    for rows.Next() {
        var game data.Game
        if err := rows.Scan(&game.ID, &game.GameName, &game.CreatorID, &game.Status, &game.CreatedAt); err != nil {
            return nil, err
        }
        games = append(games, &game)
    }
    return games, nil
}

// GetGameByID retrieves a game by its ID
func GetGameByID(gameID int) (*data.Game, error) {
    query := `SELECT id, game_name, creator_id, status, created_at FROM games WHERE id = $1`
    game := &data.Game{}
    err := DB.QueryRow(query, gameID).Scan(&game.ID, &game.GameName, &game.CreatorID, &game.Status, &game.CreatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("could not query game: %v", err)
    }
    return game, nil
}

// DeleteGame removes a game and its associated players from the database
func DeleteGame(gameID int) error {
    query := `DELETE FROM games WHERE id = $1`
    _, err := DB.Exec(query, gameID)
    if err != nil {
        return fmt.Errorf("could not delete game: %v", err)
    }
    return nil
}



// AddPlayerToGame adds a player to the players table
func AddPlayerToGame(gameID int, userID int) error {
    // Проверяем, существует ли игрок в игре
    playerExistsQuery := `SELECT 1 FROM players WHERE game_id = $1 AND user_id = $2`
    var exists bool
    err := DB.QueryRow(playerExistsQuery, gameID, userID).Scan(&exists)
    if err == sql.ErrNoRows {
        // Если игрок не существует, добавляем его
        insertPlayerQuery := `INSERT INTO players (game_id, user_id, health) VALUES ($1, $2, 4)`
        _, err = DB.Exec(insertPlayerQuery, gameID, userID)
        if err != nil {
            return fmt.Errorf("could not insert player: %v", err)
        }
    } else if err != nil {
        return fmt.Errorf("could not check player existence: %v", err)
    }

    return nil
}

// Getting players who joined the game
func GetPlayersInGame(gameID int) ([]*data.Player, error) {
    query := `
        SELECT p.id, p.game_id, p.user_id, p.health,
               COALESCE(r.name, 'No Role') AS role, 
               COALESCE(c.name, 'No Character') AS character
        FROM players p
        LEFT JOIN roles r ON p.role = r.name
        LEFT JOIN characters c ON p.character = c.name
        WHERE p.game_id = $1`

    rows, err := DB.Query(query, gameID)
    if err != nil {
        return nil, fmt.Errorf("could not query players: %v", err)
    }
    defer rows.Close()

    var players []*data.Player
    for rows.Next() {
        player := &data.Player{}
        err := rows.Scan(
            &player.ID, &player.GameID, &player.UserID, &player.Health,
            &player.Role, &player.Character,
        )
        if err != nil {
            return nil, fmt.Errorf("could not scan player: %v", err)
        }
        players = append(players, player)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("row iteration error: %v", err)
    }

    return players, nil
}

// UpdatePlayerRoleAndCharacter updates the role and character of a player
func UpdatePlayerRoleAndCharacter(playerID int, role string, character string, health int) error {
    query := `UPDATE players SET role = $1, character = $2, health = $3 WHERE id = $4`
    _, err := DB.Exec(query, role, character, health, playerID)
    if err != nil {
        log.Println("Error when UpdatePlayerRoleAndCharacter")
        return fmt.Errorf("could not update player role and character: %v", err)
    }
    return nil
}

// Dividing Roles
func GetRolesByPlayerCount(numPlayers int) ([]data.Role, error) {
    query := `SELECT name, definition FROM roles LIMIT $1`
    rows, err := DB.Query(query, numPlayers)
    if err != nil {
        log.Println("query error in roles")
        return nil, fmt.Errorf("could not query roles: %v", err)
    }
    defer rows.Close()

    var roles []data.Role
    for rows.Next() {
        var role data.Role
        err := rows.Scan(&role.Name, &role.Definition)
        if err != nil {
            log.Println("Error scanning role")
            return nil, fmt.Errorf("could not scan role: %v", err)
        }
        roles = append(roles, role)
    }

    if len(roles) < numPlayers {
        return nil, fmt.Errorf("not enough roles for %d players", numPlayers)
    }

    return roles, nil
}

// Dividing Characters
func GetAvailableCharacters(gameID int) ([]data.Character, error) {
    query := `SELECT c.name, c.definition, c.health 
              FROM characters c
              WHERE NOT EXISTS (
                  SELECT 1 FROM players p
                  WHERE p.character = c.name AND p.game_id = $1
              )`
    rows, err := DB.Query(query, gameID)
    if err != nil {
        log.Println("Error query in character")
        return nil, fmt.Errorf("could not query characters: %v", err)
    }
    defer rows.Close()

    var characters []data.Character
    for rows.Next() {
        var character data.Character
        err := rows.Scan(&character.Name, &character.Definition, &character.Health)
        if err != nil {
            log.Println("Error getting character")
            return nil, fmt.Errorf("could not scan character: %v", err)
        }
        characters = append(characters, character)
    }

    // Проверяем, что достаточно персонажей для всех игроков
    if len(characters) < gameID {
        return nil, fmt.Errorf("not enough characters available")
    }

    return characters, nil
}
