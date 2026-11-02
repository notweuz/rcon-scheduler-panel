package auth

import (
	"context"
	"errors"
	"fmt"
	"rcon-scheduler-panel/internal/auth"
	authUtil "rcon-scheduler-panel/internal/auth"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/user"
	"rcon-scheduler-panel/internal/dto/request"
	"rcon-scheduler-panel/internal/errs"
	"rcon-scheduler-panel/internal/model"
	"strings"
)

var errInvalidCredentials = errs.ErrUnauthorized.WithMessage("invalid credentials")

type service struct {
	userSVC user.Service
	cfg     *config.Config
}

var dummyHash, _ = authUtil.GeneratePasswordHash("dummy")

func NewService(userSVC user.Service, cfg *config.Config) Service {
	return &service{userSVC: userSVC, cfg: cfg}
}

func (s *service) Register(ctx context.Context, req request.AuthCredentials) (string, error) {
	if !s.cfg.AllowRegistration {
		count, err := s.userSVC.Count(ctx)
		if err != nil {
			return "", fmt.Errorf("register: count users: %w", err)
		}
		if count > 0 {
			return "", errs.ErrForbidden.WithMessage("registration is disabled")
		}
	}

	encryptedPassword, err := authUtil.GeneratePasswordHash(req.Password)
	if err != nil {
		return "", fmt.Errorf("register: hash password: %w", err)
	}
	u := &model.User{Username: strings.ToLower(req.Username), Password: encryptedPassword}
	if err := s.userSVC.Create(ctx, u); err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	token, err := authUtil.GenerateToken(u.ID, s.cfg.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("register: generate token: %w", err)
	}
	return token, nil
}

func (s *service) Login(ctx context.Context, req request.AuthCredentials) (string, error) {
	u, err := s.userSVC.GetByUsername(ctx, strings.ToLower(req.Username))
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			authUtil.VerifyPasswordHash(dummyHash, req.Password)
			return "", errInvalidCredentials.WithErr(err)
		}
		return "", fmt.Errorf("login: %w", err)
	}

	if !authUtil.VerifyPasswordHash(u.Password, req.Password) {
		return "", errInvalidCredentials
	}

	token, err := auth.GenerateToken(u.ID, s.cfg.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("login: generate token: %w", err)
	}
	return token, nil
}
