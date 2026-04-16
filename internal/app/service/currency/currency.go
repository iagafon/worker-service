package currency

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/iagafon/worker-service/internal/app/client/fixer"
	"github.com/iagafon/worker-service/internal/app/entity"
	"github.com/iagafon/worker-service/internal/app/repository"
)

// Service — сервис для работы с курсами валют.
type Service struct {
	fixerClient *fixer.Client
	rateRepo    repository.CurrencyRate
}

// NewService создаёт новый сервис курсов валют.
func NewService(fixerClient *fixer.Client, rateRepo repository.CurrencyRate) *Service {
	return &Service{
		fixerClient: fixerClient,
		rateRepo:    rateRepo,
	}
}

// GetRate возвращает курс валюты from относительно to.
func (s *Service) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	rate, err := s.rateRepo.GetRate(ctx, from, to)
	if err == nil {
		return rate, nil
	}

	rates, err := s.fixerClient.GetRates(ctx, from)
	if err != nil {
		return 0, err
	}

	if err = s.rateRepo.SetRates(ctx, from, rates); err != nil {
		log.Error().Err(err).Msg("Failed to set rates to cache")
	}

	rate, ok := rates[to]
	if !ok {
		return 0, entity.ErrFixerCurrencyNotFound
	}

	return rate, nil
}

// Convert конвертирует сумму из одной валюты в другую.
func (s *Service) Convert(ctx context.Context, amount float64, from, to string) (float64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return amount * rate, nil
}
