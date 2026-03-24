package jwt

import (
	"auth-service/domain"
	"auth-service/helper"
	"errors"
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
	issuer := viper.GetString(helper.ClientID)
	if issuer == "" {
		issuer = "auth-service"
	}

	claims := jwt.MapClaims{
		helper.TokenType: tokenType,
		helper.Exp:       expTime.Unix(),
		helper.Iss:       issuer,
		helper.Sub:       email,
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

func VerifyToken(tokenString string) (*domain.VerifyTokenResponse, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		expectedAlg := viper.GetString(helper.JWT_SIGNING_METHOD)
		if expectedAlg != "" && token.Method.Alg() != expectedAlg {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(viper.GetString(helper.JWT_SIGNATURE_KEY)), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.StatusUnauthorized, domain.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.StatusUnauthorized, domain.ErrUnauthorized
	}

	exp, ok := claims[helper.Exp].(float64)
	if !ok || int64(exp) < time.Now().Unix() {
		return nil, domain.StatusUnauthorized, domain.ErrUnauthorized
	}

	response := &domain.VerifyTokenResponse{
		Email:     getStringClaim(claims, helper.Sub),
		Subject:   getStringClaim(claims, helper.Sub),
		Issuer:    getStringClaim(claims, helper.Iss),
		TokenType: getStringClaim(claims, helper.TokenType),
	}

	return response, domain.StatusSuccess, nil
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}

	return value
}
