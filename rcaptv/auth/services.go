package auth

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/helix"
)

// Service handles related webserver services.
//
// IMPORTANT: Service must not be copied
type Service struct {
	wg             sync.WaitGroup
	TokenValidator *TokenValidator
	TokenCollector *TokenCollector
}

func (s *Service) Start() {
	l := log.With().Str("ctx", "auth_service").Logger()

	s.wg.Add(2)
	l.Info().Msg("starting TokenValidator")
	go func() {
		defer s.wg.Done()
		s.TokenValidator.Run()
	}()
	l.Info().Msg("starting TokenCollector")
	go func() {
		defer s.wg.Done()
		s.TokenCollector.Run()
	}()
}

func (s *Service) Stop() {
	l := log.With().Str("ctx", "auth_service").Logger()
	defer s.wg.Wait()
	l.Info().Msg("stopping TokenValidator")
	s.TokenValidator.Stop()
	l.Info().Msg("stopping TokenCollector")
	s.TokenCollector.Stop()
}

func NewService(sto database.Storage) *Service {
	db := sto.Conn()
	return &Service{
		TokenValidator: NewTokenValidator(db, helix.NewWithoutExchange(&helix.HelixOpts{
			// helix client used only for validating tokens
			Creds: helix.ClientCreds{
				ClientID:     "",
				ClientSecret: "",
			},
			APIUrl: "",
		})),
		TokenCollector: NewCollector(db, time.Duration(cfg.TokenCollectorIntervalHours)*time.Hour),
	}
}
