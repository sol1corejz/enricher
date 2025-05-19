package enricher

import (
	"context"
	"errors"
	"fmt"
	"github.com/sol1corejz/enricher/internal/domain/models"
	"log/slog"
)

type Enricher struct {
	log              *slog.Logger
	enricherProvider Provider
}

type Provider interface {
	SaveUser(ctx context.Context, userData models.EnrichedUser) (int64, error)
	EditUser(ctx context.Context, userData models.EnrichedUser) (models.EnrichedUser, error)
	DeleteUser(ctx context.Context, id int64) error
	GetUsers(ctx context.Context, filter models.UserFilter) ([]models.EnrichedUser, error)
	GetUser(ctx context.Context, id int64) (models.EnrichedUser, error)
}

var (
	ErrUserNotFound = errors.New("user not found")
)

// New returns a new instance of the Auth service.
func New(
	log *slog.Logger,
	enricherProvider Provider,
) *Enricher {
	return &Enricher{
		log:              log,
		enricherProvider: enricherProvider,
	}
}

func (a *Enricher) SaveUser(ctx context.Context, userData models.EnrichedUser) (int64, error) {
	const op = "enricher.SaveUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to save user")

	userID, err := a.enricherProvider.SaveUser(ctx, userData)
	if err != nil {
		a.log.Error("failed to save user", err.Error())

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (a *Enricher) EditUser(ctx context.Context, userData models.EnrichedUser) (models.EnrichedUser, error) {
	const op = "enricher.EditUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to edit user")

	user, err := a.enricherProvider.EditUser(ctx, userData)
	if err != nil {
		a.log.Error("failed to edit user", err.Error())

		return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (a *Enricher) DeleteUser(ctx context.Context, id int64) error {
	const op = "enricher.DeleteUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to delete user")

	err := a.enricherProvider.DeleteUser(ctx, id)
	if err != nil {
		a.log.Error("failed to delete user", err.Error())

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Enricher) GetUsers(ctx context.Context, filter models.UserFilter) ([]models.EnrichedUser, error) {
	const op = "enricher.GetUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to get user")

	users, err := a.enricherProvider.GetUsers(ctx, filter)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return []models.EnrichedUser{}, ErrUserNotFound
		}
		a.log.Error("failed to get user", err.Error())

		return []models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (a *Enricher) GetUser(ctx context.Context, id int64) (models.EnrichedUser, error) {
	const op = "enricher.GetUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to get user")

	user, err := a.enricherProvider.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return models.EnrichedUser{}, ErrUserNotFound
		}
		a.log.Error("failed to get user", err.Error())

		return models.EnrichedUser{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
