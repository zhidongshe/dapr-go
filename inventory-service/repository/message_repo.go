package repository

import (
	"database/sql"
	"time"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// IsProcessed checks if a message has been processed
func (r *MessageRepository) IsProcessed(messageID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM processed_messages WHERE message_id = ?",
		messageID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkProcessed marks a message as processed within a transaction
func (r *MessageRepository) MarkProcessed(tx *sql.Tx, messageID string, topic string) error {
	_, err := tx.Exec(
		"INSERT INTO processed_messages (message_id, topic, processed_at) VALUES (?, ?, ?)",
		messageID, topic, time.Now(),
	)
	return err
}

// CleanupOldMessages removes processed messages older than the specified duration
func (r *MessageRepository) CleanupOldMessages(before time.Time) error {
	_, err := r.db.Exec(
		"DELETE FROM processed_messages WHERE processed_at < ?",
		before,
	)
	return err
}
