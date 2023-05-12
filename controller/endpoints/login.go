package endpoints

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/models"
	"github.com/dgrijalva/jwt-go"
)

func generateAccessToken(userID string) (string, error) {
	// Create a new token object
	token := jwt.New(jwt.SigningMethodHS256)

	// Set token claims
	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = userID

	// Generate encoded token and return it
	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetUserIDFromAccessToken(tokenString string) (string, error) {
	// Parse the token with your JWT secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Replace "your-secret-key" with your actual secret key
		return []byte("your-secret-key"), nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %v", err)
	}

	// Extract the user ID from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to extract user ID from token")
	}
	userID, ok := claims["userID"].(string)
	if !ok {
		return "", fmt.Errorf("failed to extract user ID from token")
	}

	return userID, nil
}

func VerifyAccessToken(accessToken string) (jwt.MapClaims, error) {
	// Parse the token

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}

		// Return the secret key
		return []byte("your-secret-key"), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid-token")
	}

	// Get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Parse the request body
		var reqBody struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Query the database for the user
		var userID string
		err = db.QueryRow("SELECT id FROM users WHERE username=$1 AND password=$2", reqBody.Username, reqBody.Password).Scan(&userID)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Generate an access token for the user
		accessToken, err := generateAccessToken(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the access token in the response
		json.NewEncoder(w).Encode(struct {
			AccessToken string `json:"access_token"`
		}{AccessToken: accessToken})
	}
}
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Get the access token from the Authorization header
		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}
		// Verify the access token
		_, err := VerifyAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

//========================================================================================================================//

func GetAllUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var users []models.User

		rows, err := db.Query("select id,username,password,email from users")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var user models.User

			rowscn := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email)
			if rowscn != nil {
				log.Fatal(rowscn)
			}
			users = append(users, user)
		}
		json.NewEncoder(w).Encode(users)
	}
}

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		w.Header().Set("Content-Type", "application/json")

		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: "Error parsing the body: " + err.Error(),
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		if user.Username == "" {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: "username is required",
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		if user.Password == "" {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: "password is required",
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		err = db.QueryRowContext(ctx, "INSERT INTO users (username,password,email) VALUES ($1,$2,$3) returning id", user.Username, user.Password, user.Email).Scan(&newID)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
		}
		if err == nil {
			user = models.User{
				ID:       newID,
				Username: user.Username,
				Password: user.Password,
				Email:    user.Email,
				// Statuscode: http.StatusOK,
			}
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(user)
	}
}
