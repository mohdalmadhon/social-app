package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type payLoad struct {
	UserID     int    `json:"user_id"`
	Identifier string `json:"username"`
	Exp        int64  `json:"exp"`
	Created_at int64  `json:"iat"`
}

func GenerateToken(userID int) (string, error) {
	header := header{
		Alg: "HS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	body := payLoad{
		UserID:     userID,
		Created_at: time.Now().Unix(),
		Exp:        time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	signatureTxt, err := readTokenSecret()
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedBody := base64.RawURLEncoding.EncodeToString(bodyJSON)

	msg := encodedHeader + "." + encodedBody

	mac := hmac.New(sha256.New, signatureTxt)
	mac.Write([]byte(msg))

	signature := mac.Sum(nil)

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	token := msg + "." + encodedSignature
	return token, nil

}

func readTokenSecret() ([]byte, error) {
	if secret := os.Getenv("ORBIT_TOKEN_SECRET"); secret != "" {
		return []byte(secret), nil
	}

	return os.ReadFile("./db/tokens-signature.txt")
}

func VerifyToken(token string) (*payLoad, error) {
	parts := strings.Split(token, ".")

	if len(parts) != 3 {
		return nil, errors.New("invalid token")
	}

	encodedHeader := parts[0]
	encodedPayload := parts[1]
	encodedSignature := parts[2]

	secret, err := readTokenSecret()
	if err != nil {
		return nil, err
	}

	message := encodedHeader + "." + encodedPayload

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))

	expectedSignature := mac.Sum(nil)

	providedSignature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return nil, errors.New("invalid signature")
	}

	if !hmac.Equal(providedSignature, expectedSignature) {
		return nil, errors.New("invalid token signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, errors.New("invalid payload")
	}

	var body payLoad
	err = json.Unmarshal(payloadJSON, &body)
	if err != nil {
		return nil, errors.New("invalid payload")
	}

	if time.Now().Unix() >= body.Exp {
		return nil, errors.New("token expired")
	}

	return &body, nil
}
