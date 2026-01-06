package login

import (
	"Car-Management-System/models"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	valid := (credentials.UserName == "admin" && credentials.Password == "admin123")

	if !valid {
		http.Error(w, "Incorrect Username or Password", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateToken(credentials.UserName)
	if err != nil {
		http.Error(w, "Failed to Generate token", http.StatusInternalServerError)
		log.Println("Error Generating token: ", err)
		return
	}

	response := map[string]string{"token": tokenString}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GenerateToken(userName string) (string, error) {
	expiration := time.Now().Add(24 * time.Hour)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiration),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   userName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("some_value"))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
