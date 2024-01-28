package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"soc/pkg/domain"
)

type DialogueRepository struct {
	db *pgxpool.Pool
}

func BuildDialogueRepository(db *pgxpool.Pool) DialogueRepository {
	return DialogueRepository{db: db}
}

func (d DialogueRepository) CreateDialogue(ctx context.Context, from int, to int, message string) error {
	const query = `INSERT INTO dialogues (from_user_id, to_user_id, text) VALUES ($1, $2, $3)`
	_, err := d.db.Exec(ctx, query, from, to, message)
	if err != nil {
		return err
	}

	return nil
}

func (d DialogueRepository) GetDialogue(ctx context.Context, userId int, withUserId int) ([]domain.Dialogue, error) {
	const query = `SELECT * FROM dialogues WHERE to_user_id IN ($1,$2) AND from_user_id IN ($1,$2)`
	dialogues := make([]domain.Dialogue, 0)

	rows, err := d.db.Query(ctx, query, userId, withUserId)
	if err != nil {
		return nil, err
	}

	fmt.Println(rows)

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
