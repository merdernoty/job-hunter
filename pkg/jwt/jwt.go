package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/merdernoty/job-hunter/config"
)

type JWTService struct {
	SecretKey string
	TTL       time.Duration
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{SecretKey: cfg.Jwt.Secret, TTL: cfg.Jwt.TTL}
}

func (s *JWTService) GenerateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(s.TTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString([]byte(s.SecretKey))
}

func (s *JWTService) VerifyToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return []byte(s.SecretKey), nil

	})
	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if idStr, ok := claims["user_id"].(string); ok {
			return uuid.Parse(idStr)
		}
	}
	return uuid.Nil, jwt.ErrTokenInvalidClaims
}
