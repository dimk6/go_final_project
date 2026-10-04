package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// GenerateJWT создаёт токен HS256 с полезной нагрузкой хэша пароля
func GenerateJWT(password string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	passHash := sha256.Sum256([]byte(password))
	claimsMap := map[string]interface{}{
		"hash": base64.RawURLEncoding.EncodeToString(passHash[:]),
		"exp":  time.Now().Add(8 * time.Hour).Unix(),
	}
	claimsBytes, _ := json.Marshal(claimsMap)
	payload := base64.RawURLEncoding.EncodeToString(claimsBytes)

	data := header + "." + payload
	h := hmac.New(sha256.New, []byte(password))
	h.Write([]byte(data))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return data + "." + signature, nil
}

// ValidateJWT проверяет подпись и актуальность пароля
func ValidateJWT(tokenStr string, password string) bool {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return false
	}

	data := parts[0] + "." + parts[1]
	h := hmac.New(sha256.New, []byte(password))
	h.Write([]byte(data))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	var claims struct {
		Hash string `json:"hash"`
		Exp  int64  `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return false
	}

	if time.Now().Unix() > claims.Exp {
		return false
	}

	currHash := sha256.Sum256([]byte(password))
	if claims.Hash != base64.RawURLEncoding.EncodeToString(currHash[:]) {
		return false
	}

	return true
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	expectedPass := os.Getenv("TODO_PASSWORD")
	if expectedPass == "" || req.Password != expectedPass {
		sendJSONError(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	token, err := GenerateJWT(expectedPass)
	if err != nil {
		sendJSONError(w, "ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SignInResponse{Token: token})
}

// AuthMiddleware проверяет наличие и валидность токена, если задан TODO_PASSWORD
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if pass != "" {
			var token string
			cookie, err := r.Cookie("token")
			if err == nil {
				token = cookie.Value
			}

			if token == "" || !ValidateJWT(token, pass) {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
