package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func AddToCart(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		accessToken := r.Header.Get("Authorization")

		if accessToken == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		_, err := VerifyAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var cart models.Cart

		err = json.NewDecoder(r.Body).Decode(&cart)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if product already exists in cart
		var existingCart models.Cart
		result := db.Table("cart").Where("user_id = ? AND product_id = ?", userIDInt, cart.ProductID).First(&existingCart)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				cart.UserID = userIDInt
				cart.Quantity = 1
				cart.Final_price = cart.Final_price * cart.Quantity
				result = db.Table("cart").Create(&cart)
				if result.Error != nil {
					http.Error(w, "Failed to add to cart", http.StatusInternalServerError)
					log.Println("Error creating cart:", result.Error)
					return
				}
			} else {
				http.Error(w, "Failed to check cart", http.StatusInternalServerError)
				log.Println("Error checking cart:", result.Error)
				return
			}
		} else {
			// If the product already exists, update the quantity and final price
			existingCart.Quantity++
			existingCart.Final_price = existingCart.Quantity * cart.Final_price
			result = db.Table("cart").Save(&existingCart)
			if result.Error != nil {
				http.Error(w, "Failed to update cart", http.StatusInternalServerError)
				log.Println("Error updating cart:", result.Error)
				return
			}
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Product added to cart successfully",
		})
	}
}

func AddToCart2(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		accessToken := r.Header.Get("Authorization")

		if accessToken == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		_, err := VerifyAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var cart models.Cart

		err = json.NewDecoder(r.Body).Decode(&cart)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if product already exists in cart
		var existingCart models.Cart
		result := db.Table("cart").Where("user_id = ? AND product_id = ?", userIDInt, cart.ProductID).First(&existingCart)
		if result.Error == nil {
			// If the product already exists, update the quantity and final price
			existingCart.Quantity += 1
			existingCart.Final_price = cart.Final_price
			result := db.Table("cart").Save(&existingCart)
			if result.Error != nil {
				http.Error(w, "Failed to update cart", http.StatusInternalServerError)
				return
			}
		} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Product does not exist in cart, create a new entry
			cart.UserID = userIDInt
			result := db.Table("cart").Create(&cart)
			if result.Error != nil {
				http.Error(w, "Failed to add to cart", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to query the database", http.StatusInternalServerError)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Product added to cart successfully",
		})
	}
}

func CartItems(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		_, err := VerifyAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		id, err := GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var cart []models.Cart
		result := db.Model(&cart).
			Table("cart").
			Preload("Product").
			Where("user_id = ?", id).
			Find(&cart)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(cart)
	})
}

func ViewCartNotLoggedIn(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, _ := strconv.Atoi(params["id"])
		var product models.Product

		result := db.Table("products").Where("id = ?", id).First(&product)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				http.Error(w, "Product not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error retrieving product", http.StatusInternalServerError)
			}
			return
		}

		products := []models.Product{product}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(products)
	}
}

func UpdateQuantityAndPrice(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cart models.Cart
		json.NewDecoder(r.Body).Decode(&cart)

		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		_, err := VerifyAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		id, err := GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Update the cart using GORM
		result := db.Model(&cart).
			Table("cart").
			Where("user_id = ? AND product_id = ?", id, cart.ProductID).
			Updates(models.Cart{Quantity: cart.Quantity, Final_price: cart.Final_price})

		if result.Error != nil {
			http.Error(w, "Failed to update cart", http.StatusInternalServerError)
			return
		}

		rowsAffected := result.RowsAffected
		if rowsAffected == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not Found %s", id)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode("Success")
	}
}

func DeleteCart(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cart models.Cart
		json.NewDecoder(r.Body).Decode(&cart)

		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		_, err := VerifyAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, err := GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Delete the cart item using GORM
		result := db.Where("product_id = ? AND user_id = ?", cart.ProductID, userID).Table("cart").Delete(&models.Cart{})

		if result.Error != nil {
			http.Error(w, "Failed to delete cart item", http.StatusInternalServerError)
			return
		}

		rowsAffected := result.RowsAffected
		if rowsAffected == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not Found %s", userID)
			return
		}

		json.NewEncoder(w).Encode("Deleted Successfully!")
	}
}
