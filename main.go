package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("Прячем")

type Token struct {
	Secret   string        `json:"secret"`
	UserID   string        `json:"user_id"`
	Exp      time.Duration `json:"exp"`
	JsonLine string        `json:"json_line"`
}

type Request struct {
	Secret string `json:"secret"`
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

func checkLength(s string) bool {
	return len(s) < 500
}

func createToken(w http.ResponseWriter, r *http.Request) {

	/*
		format:
			Secret,
			UserID,
			Exp
	*/

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var readToken Token
	err := json.NewDecoder(r.Body).Decode(&readToken)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if len(readToken.JsonLine) > 5000 || readToken.UserID == "" || readToken.Exp <= 0 || readToken.Secret == "" {
		http.Error(w, "Invalid token data", http.StatusBadRequest)
		return
	}

	if !checkLength(readToken.UserID) || !checkLength(readToken.Secret) {
		http.Error(w, "Wrong length data", http.StatusBadRequest)
		return
	}

	claims := jwt.MapClaims{
		"secret":    HashString(readToken.Secret),
		"user_id":   HashString(readToken.UserID),
		"exp":       time.Now().Add(time.Hour * readToken.Exp).Unix(),
		"json_line": readToken.JsonLine,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtStr, err := token.SignedString(secret)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": jwtStr})
}

func getToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID != HashString(req.UserID) {
		http.Error(w, "Unauthorized: wrong user_id", http.StatusUnauthorized)
		return
	}

	tokenSecret, ok := claims["secret"].(string)
	if !ok || tokenSecret != HashString(req.Secret) {
		http.Error(w, "Unauthorized: wrong secret", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(claims)
}

func main() {
	r := NewRouter()

	r.Group("/token", func(api *Router) {
		api.Handle(http.MethodPost, "/create", createToken)
		api.Handle(http.MethodPost, "/get", getToken)
	})

	log.Fatal(http.ListenAndServe(":1024", r))
}
