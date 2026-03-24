package helper

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

const (
	StatusInternalServerError = "AUTH50000"
	StatusSuccessLogin        = "AUTH20001"
)

func BuildLoginResponse(id string, email string) (map[string]interface{}, string, error) {
	expSecond := time.Duration(viper.GetInt64(JWT_EXPARATION)) * time.Second
	return buildTokenResponse(id, email, "permanent", expSecond)
}

func buildTokenResponse(id string, email string, tokenType string, expSecond time.Duration) (map[string]interface{}, string, error) {
	expTime := time.Now().Add(expSecond)
	expTimeStr := expTime.Format("2006-01-02T15:04:05-0700")

	claims := jwt.MapClaims{
		IdKey:     id,
		TokenType: tokenType,
		Exp:       expTime.Unix(),
	}

	t := jwt.NewWithClaims(jwt.GetSigningMethod(viper.GetString(JWT_SIGNING_METHOD)), claims)

	token, err := t.SignedString([]byte(viper.GetString(JWT_SIGNATURE_KEY)))
	if err != nil {
		return nil, StatusInternalServerError, err
	}

	response := map[string]interface{}{
		"token":   token,
		"expired": expTimeStr,
		"email":   email,
	}
	return response, StatusSuccessLogin, nil
}
