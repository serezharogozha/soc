package repository

import (
	"dialogues/pkg/domain"
	"fmt"
	"github.com/tarantool/go-tarantool"
)

type DialogueRepository struct {
	tarantool *tarantool.Connection
}

func BuildDialogueRepository(tarantool *tarantool.Connection) DialogueRepository {
	return DialogueRepository{
		tarantool: tarantool,
	}
}

func (d DialogueRepository) CreateMessage(dialogueId uint64, from int, to int, message string) error {
	fmt.Println("CreateMessage", dialogueId, from, to, message)
	_, err := d.tarantool.Call("create_message", []interface{}{dialogueId, from, to, message})
	if err != nil {
		fmt.Println("Error in CreateMessage")
	}
	return err
}

func (d DialogueRepository) IsDialogueExist(fromUserId int, toUserId int) (uint64, error) {
	result, err := d.tarantool.Call("is_dialogue_exist", []interface{}{fromUserId, toUserId})
	if err != nil {
		fmt.Println("Error in IsDialogueExist", err)
		return 0, err
	}

	if result.Tuples()[0][0] != nil {
		dialogueId := result.Tuples()[0][0]

		return dialogueId.(uint64), nil
	}

	return 0, nil
}

func (d DialogueRepository) CreateDialogue(from int, to int) (uint64, error) {
	result, err := d.tarantool.Call("create_dialogue", []interface{}{from, to})

	if err != nil {
		fmt.Println("Error in CreateDialogue", err)
		return 0, err
	}

	dialogueId := result.Tuples()[0][0].(uint64)

	return dialogueId, nil
}

func (d DialogueRepository) ReadMessages(userId int, withUserId int) (uint64, error) {
	count, err := d.tarantool.Call("read_messages", []interface{}{userId, withUserId})
	if err != nil {
		fmt.Println("Error in ReadMessages", err)
		return 0, err
	}
	fmt.Println("ReadMessages", count.Data[0])

	counter := count.Tuples()[0][0].(uint64)

	return counter, nil
}

func (d DialogueRepository) UnreadMessages(userId int, withUserId int, counter uint64) {
	err, _ := d.tarantool.Call("unread_messages", []interface{}{userId, withUserId, counter})
	if err != nil {
		return
	}
	return
}

func (d DialogueRepository) GetDialogue(userId int, withUserId int) ([]domain.Dialogue, error) {
	result, err := d.tarantool.Call("get_dialogue", []interface{}{userId, withUserId})
	if err != nil {
		fmt.Println("Error in GetDialogue", err)
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
