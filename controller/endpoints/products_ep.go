package endpoints

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/models"
	"github.com/gorilla/mux"
)

func CreateProduct(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// err := os.Mkdir("uploads", 0755)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		w.Header().Set("Content-Type", "application/json")
		err := r.ParseMultipartForm(32 << 20) // max memory 32MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		name := r.FormValue("name")
		category_id := r.FormValue("category_id")
		image, handler, err := r.FormFile("image")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error getting image file: %v", err)
			return
		}
		description := r.FormValue("description")
		seller := r.FormValue("seller")
		pricestr := r.FormValue("price")
		highlights := r.FormValue("highlights")
		specifications := r.FormValue("specifications")

		price, err := strconv.Atoi(pricestr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error getting image file: %v", err)
			return
		}

		defer image.Close()
		allowedExtensions := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}

		filename := handler.Filename
		ext := filepath.Ext(filename)
		if !allowedExtensions[ext] {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Invalid file format. Only JPG, JPEG, PNG files are allowed.")
			return
		}

		tempFile, err := os.CreateTemp("", "upload-*"+ext)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error creating temporary file: %v", err)
			return
		}

		defer tempFile.Close()
		io.Copy(tempFile, image)

		imageURL := tempFile.Name()
		filepath := fmt.Sprintf("../uploads/%s", handler.Filename)
		err = os.Rename(tempFile.Name(), filepath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error moving file to uploads directory: %v", err)
			return
		}

		imageURL = filepath

		//the newPath variable is created to hold the file path of the uploaded file
		// in the "uploads" directory. filepath.Join is used to join the "uploads" directory and
		// the file name to create the complete file path.
		// newPath := filepath.Join("uploads", filename)
		// err = os.Rename(tempFile.Name(), newPath)
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	fmt.Fprintf(w, "Error moving file to uploads directory: %v", err)
		// 	return
		// }
		// portwithoutcolun := strings.Replace(port, ":", "", 1)
		// imageURL := fmt.Sprintf("http://localhost:"+portwithoutcolun+"/uploads/%s", filename)
		// imageURL = fmt.Sprintf("http://%s/uploads/%s", portwithoutcolun, filename)
		// if category == "" {
		// 	apierror := models.APIError{
		// 		Code:    http.StatusBadRequest,
		// 		Message: "Category name is required",
		// 	}
		// 	w.WriteHeader(apierror.Code)
		// 	json.NewEncoder(w).Encode(apierror)
		// 	return
		// }
		stmt, err := db.Prepare(`INSERT INTO products (name,category_id,image_url,description,seller,price,highlights,specifications ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w, "Error preparing SQL statement: %v", err)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(name, category_id, imageURL, description, seller, price, highlights, specifications)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error executing SQL statement: %v", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Product added successfully")
		json.NewEncoder(w).Encode("added")
	}
}

func GetAllProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var products []models.Product
		// var row_var error

		rows, err := db.Query(`SELECT p.id,p.name,p.category_id,p.image_url,p.description,p.seller,p.price,p.highlights,p.specifications
								,c.category FROM products p left join category c on(p.category_id=c.id)`)

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var product models.Product
			row_var := rows.Scan(&product.ID, &product.Name, &product.CategoryID, &product.ImageURL, &product.Description, &product.Seller, &product.Price, &product.Highlights, &product.Specifications,
				&product.Category.Category)
			if row_var != nil {
				log.Fatal(row_var)
			}
			products = append(products, product)
		}
		json.NewEncoder(w).Encode(products)
	}
}
func GetAllProductsByCID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var products []models.Product
		// var row_var error
		param := mux.Vars(r)
		id, _ := strconv.Atoi(param["id"])

		rows, err := db.Query(`SELECT p.id,p.name,p.category_id,p.image_url,p.description,p.seller,p.price,p.highlights,p.specifications
								,c.category FROM products p left join category c on(p.category_id=c.id) where p.category_id=$1`, id)

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}
		defer rows.Close()

		for rows.Next() {

			var product models.Product
			err := rows.Scan(&product.ID, &product.Name, &product.CategoryID, &product.ImageURL, &product.Description, &product.Seller, &product.Price, &product.Highlights, &product.Specifications,
				&product.Category.Category)
			if err != nil {
				apierror := models.APIError{
					Code:    http.StatusBadRequest,
					Message: "No such rows with provided id is available!",
				}
				w.WriteHeader(apierror.Code)
				json.NewEncoder(w).Encode(apierror)
				return
			}
			products = append(products, product)
		}
		json.NewEncoder(w).Encode(products)
	}
}

// func SearchProducts(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		query := r.URL.Query().Get("query")

// 		// Prepare the SQL statement with conditional logic based on the search criteria
// 		stmt, err := db.Prepare(`
// 		SELECT p.id, p.name, p.category_id, p.description, p.image_url, p.seller, p.price, p.highlights, p.specifications,
// 		c.id AS category_id, c.category, c.imageurl AS category_imageurl
// 		FROM products AS p
// 		INNER JOIN category AS c ON p.category_id = c.id
// 		WHERE
// 			REPLACE(LOWER(c.category), ' ', '') LIKE LOWER($1)
// 			OR REPLACE(LOWER(p.name), ' ', '') LIKE LOWER($2)
// 		`)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer stmt.Close()

// 		// Execute the SQL statement with the appropriate parameters based on the search criteria
// 		queryParam := "%" + query + "%"
// 		rows, err := stmt.Query(queryParam, queryParam)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer rows.Close()

// 		var products []models.Product
// 		for rows.Next() {
// 			var product models.Product
// 			var category models.Category
// 			err := rows.Scan(
// 				&product.ID, &product.Name, &product.CategoryID, &product.Description, &product.ImageURL,
// 				&product.Seller, &product.Price, &product.Highlights, &product.Specifications,
// 				&category.ID, &category.Category, &category.ImageURL,
// 			)

// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			product.Category = category
// 			products = append(products, product)
// 		}
// 		if err = rows.Err(); err != nil {
// 			log.Fatal(err)
// 		}

// 		jsonData, err := json.Marshal(products)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

//			w.Header().Set("Content-Type", "application/json")
//			w.Write(jsonData)
//		}
//	}
func SearchProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")

		// Remove spaces from the search query
		query = strings.ReplaceAll(query, " ", "")

		// Prepare the SQL statement with improved search logic
		stmt, err := db.Prepare(`
			SELECT p.id, p.name, p.category_id, p.description, p.image_url, p.seller, p.price, p.highlights, p.specifications,
			c.id AS category_id, c.category, c.imageurl AS category_imageurl
			FROM products AS p
			INNER JOIN category AS c ON p.category_id = c.id
			WHERE
				REPLACE(LOWER(c.category), ' ', '') ILIKE $1
				OR REPLACE(LOWER(p.name), ' ', '') ILIKE $2
		`)
		if err != nil {
			log.Fatal(err)
		}
		defer stmt.Close()

		// Execute the SQL statement with the appropriate parameters based on the search criteria
		queryParam := "%" + query + "%"
		rows, err := stmt.Query(queryParam, queryParam)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var products []models.Product
		for rows.Next() {
			var product models.Product
			var category models.Category
			err := rows.Scan(
				&product.ID, &product.Name, &product.CategoryID, &product.Description, &product.ImageURL,
				&product.Seller, &product.Price, &product.Highlights, &product.Specifications,
				&category.ID, &category.Category, &category.ImageURL,
			)

			if err != nil {
				log.Fatal(err)
			}
			product.Category = category
			products = append(products, product)
		}
		if err = rows.Err(); err != nil {
			log.Fatal(err)
		}

		jsonData, err := json.Marshal(products)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}
