package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"soc/pkg/domain"
)

type DialogueRepository struct {
	db *pgxpool.Pool
}

func BuildDialogueRepository(db *pgxpool.Pool) DialogueRepository {
	return DialogueRepository{db: db}
}

func (d DialogueRepository) CreateDialogue(ctx context.Context, from int, to int) (int, error) {
	var dialogueId int
	const query = `INSERT INTO dialogues (from_user_id, to_user_id) VALUES ($1, $2) RETURNING id`
	err := d.db.QueryRow(ctx, query, from, to).Scan(&dialogueId)
	if err != nil {
		return 0, err
	}

	return dialogueId, nil
}

func (d DialogueRepository) CreateMessage(ctx context.Context, dialogue_id int, from int, to int, message string) error {
	const query = `INSERT INTO messages (dialogue_id, from_user_id, to_user_id, text) VALUES ($1, $2, $3, $4)`
	_, err := d.db.Exec(ctx, query, dialogue_id, from, to, message)
	if err != nil {
		return err
	}

	return nil
}

func (d DialogueRepository) GetDialogue(ctx context.Context, userId int, withUserId int) ([]domain.Dialogue, error) {
	var dialogueId int
	const dialogueQuery = `SELECT id FROM dialogues WHERE 
                            (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)`

	err := d.db.QueryRow(ctx, dialogueQuery, userId, withUserId).Scan(&dialogueId)

	if err != nil {
		return nil, err
	}

	const messageQuery = `SELECT from_user_id, to_user_id, text FROM messages WHERE dialogue_id = $1`

	rows, err := d.db.Query(ctx, messageQuery, dialogueId)
	if err != nil {
		return nil, err
	}

	dialogues := make([]domain.Dialogue, 0)

	for rows.Next() {
		dialogue := new(domain.Dialogue)
		err := rows.Scan(&dialogue.From, &dialogue.To, &dialogue.Text)
		if err != nil {
			return nil, err
		}

		dialogues = append(dialogues, *dialogue)
	}

	return dialogues, nil
}

func (d DialogueRepository) IsDialogueExist(ctx context.Context, fromUserId int, toUserId int) (int, error) {
	var dialogueId int
	const query = `SELECT id FROM dialogues WHERE 
                            (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)`

	err := d.db.QueryRow(ctx, query, fromUserId, toUserId).Scan(&dialogueId)

	if err != nil {
		return 0, err
	}

	return dialogueId, nil
}
