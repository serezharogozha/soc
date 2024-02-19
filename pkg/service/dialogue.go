package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"soc/pkg/domain"
	"soc/pkg/repository"
)

type DialogueService struct {
	dialogueRepository repository.DialogueRepository
}

func BuildDialogueService(dialogueRepository repository.DialogueRepository) DialogueService {
	return DialogueService{
		dialogueRepository: dialogueRepository,
	}
}

func (ds DialogueService) CreateMessages(ctx context.Context, from int, to int, message string) error {
	dialogueId, err := ds.dialogueRepository.IsDialogueExist(ctx, from, to)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if dialogueId == 0 {
		dialogueId, err = ds.dialogueRepository.CreateDialogue(ctx, from, to)
		if err != nil {
			return err
		}
	}

	return ds.dialogueRepository.CreateMessage(ctx, dialogueId, from, to, message)
}

func (ds DialogueService) GetDialogue(ctx context.Context, userId int, withUserId int) ([]domain.Dialogue, error) {
	dialogues, err := ds.dialogueRepository.GetDialogue(ctx, userId, withUserId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return dialogues, nil
}

func (ds DialogueService) CreateMessagesV2(from int, to int, message string) error {
	dialogueId, err := ds.dialogueRepository.IsDialogueExistV2(from, to)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if dialogueId == 0 {
		dialogueId, err = ds.dialogueRepository.CreateDialogueV2(from, to)
		if err != nil {
			return err
		}
	}

	err = ds.dialogueRepository.CreateMessageV2(dialogueId, from, to, message)

	if err != nil {
		return err
	}
	return nil
}

func (ds DialogueService) GetDialogueV2(userId int, withUserId int) ([]domain.Dialogue, error) {
	dialogues, err := ds.dialogueRepository.GetDialogueV2(userId, withUserId)
	if err != nil {
		return nil, err
	}

	return dialogues, nil
}
