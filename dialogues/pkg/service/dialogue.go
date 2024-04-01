package service

import (
	"dialogues/pkg/domain"
	"dialogues/pkg/repository"
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

func (ds DialogueService) GetDialogue(userId int, withUserId int) ([]domain.Dialogue, uint64, error) {
	dialogues, err := ds.dialogueRepository.GetDialogue(userId, withUserId)
	if err != nil {
		return nil, 0, err
	}

	readCounter, err := ds.dialogueRepository.ReadMessages(userId, withUserId)

	if err != nil {
		return nil, 0, err
	}

	return dialogues, readCounter, nil
}

func (ds DialogueService) UndoDecrement(userId int, withUserId int, count uint64) {
	ds.dialogueRepository.UnreadMessages(userId, withUserId, count)
}
