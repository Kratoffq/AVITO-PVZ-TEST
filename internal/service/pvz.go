package service

import (
	"github.com/avito/pvz/internal/repository"
)

type PVZService struct {
	repo *repository.PVZRepository
}

func NewPVZService(repo *repository.PVZRepository) *PVZService {
	return &PVZService{
		repo: repo,
	}
}

// TODO: Implement service methods
