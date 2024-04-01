package repository

import (
	"github.com/tarantool/go-tarantool"
)

type CounterRepository struct {
	tarantool *tarantool.Connection
}

func BuildCounterRepository(tarantool *tarantool.Connection) CounterRepository {
	return CounterRepository{
		tarantool: tarantool,
	}
}

func (d CounterRepository) IncrementCounter(userId int, withUserId int) error {
	_, err := d.tarantool.Call("increment_counter", []interface{}{userId, withUserId})
	return err
}

func (d CounterRepository) GetCounter(userId int, withUserId int) (uint64, error) {
	result, err := d.tarantool.Call("get_unread_counter", []interface{}{withUserId, userId})
	if err != nil {
		return 0, err
	}

	counter := result.Tuples()[0][3].(uint64)

	return counter, nil
}

func (d CounterRepository) DecrementCounter(userId int, withUserId int, counter uint64) error {
	_, err := d.tarantool.Call("decrement_counter", []interface{}{userId, withUserId, counter})
	if err != nil {
		return err
	}

	return nil
}
