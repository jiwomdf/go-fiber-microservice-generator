package jwt

import (
	"auth-service/domain"
	"auth-service/helper"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

func BuildLoginResponse(email string) (*domain.LoginResponse, string, error) {
	expSecond := time.Duration(viper.GetInt64(helper.JWT_EXPARATION)) * time.Second
	return buildTokenResponse(email, "permanent", expSecond)
}

func buildTokenResponse(email string, tokenType string, expSecond time.Duration) (*domain.LoginResponse, string, error) {
	expTime := time.Now().Add(expSecond)
	expTimeStr := expTime.Format("2006-01-02T15:04:05-0700")

	claims := jwt.MapClaims{
		helper.TokenType: tokenType,
		helper.Exp:       expTime.Unix(),
	}

	t := jwt.NewWithClaims(jwt.GetSigningMethod(viper.GetString(helper.JWT_SIGNING_METHOD)), claims)

	token, err := t.SignedString([]byte(viper.GetString(helper.JWT_SIGNATURE_KEY)))
	if err != nil {
		return nil, domain.StatusInternalServerError, err
	}

	response := domain.LoginResponse{
		Token:   token,
		Expired: expTimeStr,
		Email:   email,
	}
	return &response, domain.StatusSuccessLogin, nil
}
