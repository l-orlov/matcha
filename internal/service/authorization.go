package service

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/l-orlov/matcha/internal/config"
	"github.com/l-orlov/matcha/internal/models"
	"github.com/l-orlov/matcha/internal/repository"
	"github.com/pkg/errors"
)

type (
	AuthorizationService struct {
		cfg  *config.Config
		repo repository.SessionCache
	}
)

func NewAuthorizationService(cfg *config.Config, repo *repository.Repository) *AuthorizationService {
	return &AuthorizationService{
		cfg:  cfg,
		repo: repo,
	}
}

func (s *AuthorizationService) CreateSession(userID, fingerprint string) (accessToken, refreshToken string, err error) {
	accessTokenID := uuid.New().String()
	accessToken, err = newToken(
		userID, accessTokenID, s.cfg.JWT.SigningKey, s.cfg.JWT.AccessTokenLifetime.Duration(),
	)
	if err != nil {
		return "", "", err
	}

	refreshToken = uuid.New().String()

	err = s.repo.PutSessionAndAccessToken(models.Session{
		UserID:        userID,
		AccessTokenID: accessTokenID,
		Fingerprint:   fingerprint,
	}, refreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthorizationService) ValidateAccessToken(accessToken string) (*jwt.StandardClaims, error) {
	accessTokenClaims, err := validateToken(accessToken, s.cfg.JWT.SigningKey)
	if err != nil {
		return nil, err
	}

	// check accessToken is active
	if _, err := s.repo.GetAccessTokenData(accessTokenClaims.Id); err != nil {
		return nil, errors.Wrap(err, "not active accessToken")
	}

	return accessTokenClaims, nil
}

func (s *AuthorizationService) RefreshSession(
	currentRefreshToken, fingerprint string,
) (accessToken, refreshToken string, err error) {
	session, err := s.repo.GetSession(currentRefreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, "session not found")
	}

	if err = s.repo.DeleteSession(currentRefreshToken); err != nil {
		return "", "", err
	}

	if err = s.repo.DeleteUserToSession(session.UserID, currentRefreshToken); err != nil {
		return "", "", err
	}

	if err = s.repo.DeleteAccessToken(session.AccessTokenID); err != nil {
		return "", "", err
	}

	if session.Fingerprint != fingerprint {
		return "", "", errors.New("fingerprint does not match current one")
	}

	return s.CreateSession(session.UserID, fingerprint)
}

func (s *AuthorizationService) RevokeSession(accessToken string) error {
	accessTokenClaims, err := validateToken(accessToken, s.cfg.JWT.SigningKey)
	if err != nil {
		return err
	}

	refreshToken, err := s.repo.GetAccessTokenData(accessTokenClaims.Id)
	if err != nil {
		return errors.Wrap(err, "not active accessToken")
	}

	if err := s.repo.DeleteAccessToken(accessTokenClaims.Id); err != nil {
		return err
	}

	session, _ := s.repo.GetSession(refreshToken)
	if session != nil {
		if err = s.repo.DeleteUserToSession(session.UserID, refreshToken); err != nil {
			return err
		}
	}

	if err := s.repo.DeleteSession(refreshToken); err != nil {
		return err
	}

	return nil
}

func newToken(userID, tokenID, signingKey string, lifetime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		Id:        tokenID,
		NotBefore: time.Now().Unix(),
		ExpiresAt: time.Now().Add(lifetime).Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   userID,
	})

	return token.SignedString([]byte(signingKey))
}

func validateToken(token string, signingKey string) (*jwt.StandardClaims, error) {
	claims, err := getTokenClaims(token, signingKey)
	if err != nil {
		return nil, errors.Wrap(err, "not valid token")
	}

	// check accessToken has not expired
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("expired accessToken")
	}

	return claims, nil
}

func getTokenClaims(tokenString string, signingKey string) (*jwt.StandardClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(*jwt.StandardClaims), nil
}