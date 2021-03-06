package models

import (
	"database/sql"
	"time"
)

type ChatLog struct{}

func (*ChatLog) Insert(userID int, message string, room string) error {
	var stmt *sql.Stmt
	if v, err := db.Prepare(`
		INSERT INTO chat_log (user_id, message, room)
		VALUES (?, ?, ?)
	`); err != nil {
		return err
	} else {
		stmt = v
	}
	defer stmt.Close() // nolint: errcheck

	_, err := stmt.Exec(userID, message, room)
	return err
}

func (*ChatLog) InsertDiscord(discordName string, message string) error {
	var stmt *sql.Stmt
	if v, err := db.Prepare(`
		INSERT INTO chat_log (user_id, discord_name, message, room)
		VALUES (0, ?, ?, "lobby")
	`); err != nil {
		return err
	} else {
		stmt = v
	}
	defer stmt.Close() // nolint: errcheck

	_, err := stmt.Exec(discordName, message)
	return err
}

type ChatMessage struct {
	Name        string         `json:"name"`
	DiscordName sql.NullString `json:"discordName"`
	Message     string         `json:"message"`
	Datetime    time.Time      `json:"datetime"`
}

// Get the past messages sent in the lobby
func (*ChatLog) Get(room string, count int) ([]ChatMessage, error) {
	var rows *sql.Rows
	if v, err := db.Query(`
		SELECT
			IFNULL(users.username, "__server"),
			chat_log.discord_name,
			chat_log.message,
			chat_log.datetime_sent
		FROM
			chat_log
		LEFT JOIN
			users ON users.id = chat_log.user_id
		WHERE
			room = ?
		ORDER BY
			chat_log.datetime_sent DESC
		LIMIT
			?
	`, room, count); err != nil {
		return nil, err
	} else {
		rows = v
	}
	defer rows.Close() // nolint: errcheck

	chatMessages := make([]ChatMessage, 0)
	for rows.Next() {
		var message ChatMessage
		if err := rows.Scan(
			&message.Name,
			&message.DiscordName,
			&message.Message,
			&message.Datetime,
		); err != nil {
			return nil, err
		}
		chatMessages = append(chatMessages, message)
	}

	return chatMessages, nil
}
