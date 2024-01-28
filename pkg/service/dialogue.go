package service

import (
	"context"
	"fmt"
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

func (ds DialogueService) CreateDialogue(ctx context.Context, from int, to int, message string) error {
	return ds.dialogueRepository.CreateDialogue(ctx, from, to, message)
}

func (ds DialogueService) GetDialogue(ctx context.Context, userId int, withUserId int) ([]domain.Dialogue, error) {
	dialogues, err := ds.dialogueRepository.GetDialogue(ctx, userId, withUserId)
	fmt.Println(dialogues)
	if err != nil {
		return nil, err
	}

	return dialogues, nil
}
