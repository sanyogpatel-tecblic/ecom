package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/models"
	"github.com/gorilla/mux"
)

func AddToCart(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		var cart models.Cart

		err = json.NewDecoder(r.Body).Decode(&cart)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if product already exists in cart
		var quantity int
		err = db.QueryRow("SELECT quantity FROM cart WHERE user_id = $1 AND product_id = $2", id, cart.ProductID).Scan(&quantity)

		if err == nil {
			// If the product already exists, update the quantity and final price
			_, err = db.Exec("UPDATE cart SET quantity = $1, final_price = $2 WHERE user_id = $3 AND product_id = $4", quantity+1, (quantity+1)*cart.Final_price, id, cart.ProductID)
			if err != nil {
				http.Error(w, "Failed to update cart", http.StatusInternalServerError)
				return
			}
		} else {
			_, err = db.Exec("INSERT INTO cart (user_id, product_id, quantity, final_price) VALUES ($1, $2, $3, $4)", id, cart.ProductID, 1, cart.Final_price)
			if err != nil {
				http.Error(w, "Failed to add to cart", http.StatusInternalServerError)
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

func AddToCart2(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		var cart models.Cart

		err = json.NewDecoder(r.Body).Decode(&cart)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check if product already exists in cart
		var quantity int
		err = db.QueryRow("SELECT quantity FROM cart WHERE user_id = $1 AND product_id = $2", id, cart.ProductID).Scan(&quantity)
		log.Println(quantity)
		if err == nil {
			// If the product already exists, update the quantity and final price
			_, err = db.Exec("UPDATE cart SET quantity = $1, final_price = $2 WHERE user_id = $3 AND product_id = $4", quantity+1, cart.Final_price, id, cart.ProductID)
			if err != nil {
				http.Error(w, "Failed to update cart", http.StatusInternalServerError)
				return
			}
		} else {
			_, err = db.Exec("INSERT INTO cart (user_id, product_id, quantity, final_price) VALUES ($1, $2, $3, $4)", id, cart.ProductID, cart.Quantity, cart.Final_price)
			if err != nil {
				http.Error(w, "Failed to add to cart", http.StatusInternalServerError)
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

func CartItems(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("content-type", "application/json")
		var carts []models.Cart
		json.NewDecoder(r.Body).Decode(&carts)
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

		rows, err := db.Query(`select c.id,c.product_id,c.user_id,c.quantity,c.final_price, p.id,p.name,p.image_url,p.description,p.seller,p.price,p.highlights,p.specifications from
							cart c left join products p on(c.product_id=p.id) where c.user_id=$1`, id)

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var cart models.Cart

			err := rows.Scan(&cart.ID, &cart.ProductID, &cart.UserID, &cart.Quantity, &cart.Final_price, &cart.Product.ID, &cart.Product.Name, &cart.Product.ImageURL, &cart.Product.Description,
				&cart.Product.Seller, &cart.Product.Price, &cart.Product.Highlights, &cart.Product.Specifications)

			if err != nil {
				log.Println(err)
			}
			carts = append(carts, cart)
		}
		json.NewEncoder(w).Encode(carts)
	}
}

func ViewCartNotLoggedIn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		param := mux.Vars(r)
		id, _ := strconv.Atoi(param["id"])
		var products []models.Product

		rows, err := db.Query("select id,name,image_url,description,seller,price,highlights,specifications from products where id=$1", id)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		var product models.Product
		for rows.Next() {

			err = rows.Scan(
				&product.ID,
				&product.Name,
				&product.ImageURL,
				&product.Description,
				&product.Seller,
				&product.Price,
				&product.Highlights,
				&product.Specifications,
			)
			products = append(products, product)
		}
		if err != nil {
			log.Println(err)
		}
		if err == nil {
			product = models.Product{
				ID:             product.ID,
				Name:           product.Name,
				ImageURL:       product.ImageURL,
				Description:    product.Description,
				Seller:         product.Seller,
				Price:          product.Price,
				Highlights:     product.Highlights,
				Specifications: product.Specifications,
			}
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(products)
	}
}

func UpdatequantityAndPrice(db *sql.DB) http.HandlerFunc {
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
		// var quantity int

		result, err := db.Exec("UPDATE cart SET quantity = $1, final_price=$2 WHERE user_id = $3 AND product_id = $4", cart.Quantity, cart.Final_price, id, cart.ProductID)

		if err != nil {
			http.Error(w, "Failed to update cart", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error getting rows affected: %s", err.Error())
			return
		}
		if rowsAffected == 0 {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Not Found %s", id)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode("Success")
	}
}

func DeleteCart(db *sql.DB) http.HandlerFunc {
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
		user_id, err := GetUserIDFromAccessToken(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		_, err = db.Exec("delete from cart where product_id=$1 AND user_id=$2", cart.ProductID, user_id)

		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(w).Encode("Deleted Successfully!")
	}
}
