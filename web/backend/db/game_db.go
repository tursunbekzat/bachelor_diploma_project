package db

import (
	"backend/data"
	"database/sql"
	"fmt"
	"log"
	"time"
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

// GetAllGames retrieves all games along with their creators
func GetAllGames() ([]map[string]interface{}, error) {
    query := `
        SELECT g.id, g.game_name, g.creator_id, g.status, g.created_at, u.username 
        FROM games g
        JOIN users u ON g.creator_id = u.id
    `
    rows, err := DB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var games []map[string]interface{}
    for rows.Next() {
        var gameID int
        var gameName, status, creatorName string
        var creatorID int
        var createdAt time.Time

        err := rows.Scan(&gameID, &gameName, &creatorID, &status, &createdAt, &creatorName)
        if err != nil {
            return nil, err
        }

        game := map[string]interface{}{
            "id":           gameID,
            "game_name":    gameName,
            "creator_id":   creatorID,
            "creator_name": creatorName,
            "status":       status,
            "created_at":   createdAt,
        }
        games = append(games, game)
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
// GetPlayersInGame retrieves players in a game with their usernames
func GetPlayersInGame(gameID int) ([]map[string]interface{}, error) {
    query := `
        SELECT p.id, p.user_id, p.game_id, u.username, p.health,
               COALESCE(r.name, 'No Role') AS role, 
               COALESCE(c.name, 'No Character') AS character
        FROM players p
        JOIN users u ON p.user_id = u.id
        LEFT JOIN roles r ON p.role = r.name
        LEFT JOIN characters c ON p.character = c.name
        WHERE p.game_id = $1
    `

    rows, err := DB.Query(query, gameID)
    if err != nil {
        return nil, fmt.Errorf("could not query players: %v", err)
    }
    defer rows.Close()

    var players []map[string]interface{}
    for rows.Next() {
        var id, userID, gameID, health int
        var username, role, character string

        err := rows.Scan(&id, &userID, &gameID, &username, &health, &role, &character)
        if err != nil {
            return nil, fmt.Errorf("could not scan player: %v", err)
        }

        player := map[string]interface{}{
            "id":        id,
            "user_id":   userID,
            "game_id":   gameID,
            "username":  username,
            "health":    health,
            "role":      role,
            "character": character,
        }

        players = append(players, player)
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

func GetAvailableCharacters(gameID int, numPlayers int) ([]data.Character, error) {
    query := `SELECT c.name, c.definition, c.health 
              FROM characters c
              WHERE NOT EXISTS (
                  SELECT 1 FROM players p
                  WHERE p.character = c.name AND p.game_id = $1
              )`
    rows, err := DB.Query(query, gameID)
    if err != nil {
        log.Println("Error querying characters")
        return nil, fmt.Errorf("could not query characters: %v", err)
    }
    defer rows.Close()

    var characters []data.Character
    for rows.Next() {
        var character data.Character
        err := rows.Scan(&character.Name, &character.Definition, &character.Health)
        if err != nil {
            log.Println("Error scanning character")
            return nil, fmt.Errorf("could not scan character: %v", err)
        }
        characters = append(characters, character)
    }

    log.Println("Characters available:", len(characters))

    // Проверяем, что достаточно персонажей для всех игроков
    if len(characters) < numPlayers {
        return nil, fmt.Errorf("not enough characters available")
    }

    return characters, nil
}




// CheckPlayerExists проверяет, существует ли игрок в игре
func CheckPlayerExists(gameID int, userID int) (bool, error) {
    var exists bool
    query := `SELECT 1 FROM players WHERE game_id = $1 AND user_id = $2`
    err := DB.QueryRow(query, gameID, userID).Scan(&exists)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, fmt.Errorf("could not check player existence: %v", err)
    }
    return exists, nil
}