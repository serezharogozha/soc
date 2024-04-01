package service

import (
	"counter/pkg/repository"
)

type CounterService struct {
	counterRepository repository.CounterRepository
}

func BuildCounterService(counterRepository repository.CounterRepository) CounterService {
	return CounterService{
		counterRepository: counterRepository,
	}
}

func (cs CounterService) IncrementCounter(userId int, withUserId int) error {
	return cs.counterRepository.IncrementCounter(userId, withUserId)
}

func (cs CounterService) GetCounter(userId int, withUserId int) (uint64, error) {
	return cs.counterRepository.GetCounter(userId, withUserId)
}

func (cs CounterService) DecrementCounter(userId int, withUserId int, counter uint64) error {
	return cs.counterRepository.DecrementCounter(userId, withUserId, counter)
}
