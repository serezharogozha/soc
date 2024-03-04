package service

import (
	"dialogues/pkg/domain"
	"dialogues/pkg/repository"
	"fmt"
)

type DialogueService struct {
	dialogueRepository repository.DialogueRepository
}

func BuildDialogueService(dialogueRepository repository.DialogueRepository) DialogueService {
	return DialogueService{
		dialogueRepository: dialogueRepository,
	}
}

func (ds DialogueService) CreateMessages(from int, to int, message string) error {
	dialogueId, err := ds.dialogueRepository.IsDialogueExist(from, to)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if dialogueId == 0 {
		dialogueId, err = ds.dialogueRepository.CreateDialogue(from, to)
		if err != nil {
			return err
		}
	}

	err = ds.dialogueRepository.CreateMessage(dialogueId, from, to, message)

	if err != nil {
		return err
	}
	return nil
}

func (ds DialogueService) GetDialogue(userId int, withUserId int) ([]domain.Dialogue, error) {
	dialogues, err := ds.dialogueRepository.GetDialogue(userId, withUserId)
	if err != nil {
		return nil, err
	}

	return dialogues, nil
}
