package repository

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/tarantool/go-tarantool"
	"soc/pkg/domain"
)

type DialogueRepository struct {
	db        *pgxpool.Pool
	tarantool *tarantool.Connection
}

func BuildDialogueRepository(db *pgxpool.Pool, tarantool *tarantool.Connection) DialogueRepository {
	return DialogueRepository{
		db:        db,
		tarantool: tarantool,
	}
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

func (d DialogueRepository) CreateMessage(ctx context.Context, dialogueId int, from int, to int, message string) error {
	const query = `INSERT INTO messages (dialogue_id, from_user_id, to_user_id, text) VALUES ($1, $2, $3, $4)`
	_, err := d.db.Exec(ctx, query, dialogueId, from, to, message)
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

func (d DialogueRepository) CreateDialogueV2(from int, to int) (uint64, error) {
	result, err := d.tarantool.Call("create_dialogue", []interface{}{from, to})

	if err != nil {
		return 0, err
	}

	dialogueId := result.Tuples()[0][0].(uint64)

	return dialogueId, nil
}

func (d DialogueRepository) CreateMessageV2(dialogueId uint64, from int, to int, message string) error {
	_, err := d.tarantool.Call("create_message", []interface{}{dialogueId, from, to, message})
	return err
}

func (d DialogueRepository) IsDialogueExistV2(fromUserId int, toUserId int) (uint64, error) {
	result, err := d.tarantool.Call("is_dialogue_exist", []interface{}{fromUserId, toUserId})
	if err != nil {
		return 0, err
	}

	if result.Tuples()[0][0] != nil {
		dialogueId := result.Tuples()[0][0]

		return dialogueId.(uint64), nil
	}

	return 0, nil
}

func (d DialogueRepository) GetDialogueV2(userId int, withUserId int) ([]domain.Dialogue, error) {
	result, err := d.tarantool.Call("get_dialogue", []interface{}{userId, withUserId})
	if err != nil {
		return nil, err
	}

	dialogues := make([]domain.Dialogue, 0)

	for _, tuple := range result.Tuples() {
		if tuple[0] == nil {
			continue
		}

		dialogue := domain.Dialogue{
			From: int(tuple[2].(uint64)),
			To:   int(tuple[3].(uint64)),
			Text: tuple[4].(string),
		}
		dialogues = append(dialogues, dialogue)
	}

	return dialogues, nil
}
