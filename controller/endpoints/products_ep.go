package endpoints

import (
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
	"gorm.io/gorm"
)

func CreateProduct(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := r.ParseMultipartForm(32 << 20) // max memory 32MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		name := r.FormValue("name")
		categoryID, err := strconv.Atoi(r.FormValue("category_id"))
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

		product := models.Product{
			Name:           name,
			CategoryID:     categoryID,
			ImageURL:       imageURL,
			Description:    description,
			Seller:         seller,
			Price:          price,
			Highlights:     highlights,
			Specifications: specifications,
		}

		result := db.Create(&product)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error creating product: %v", result.Error)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Product added successfully")
		json.NewEncoder(w).Encode("added")
	}
}

func GetAllProducts(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var products []models.Product

		result := db.Find(&products)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(products)
	})
}

func GetAllProductsByCID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var products []models.Product

		param := mux.Vars(r)
		id, _ := strconv.Atoi(param["id"])

		result := db.Where("category_id = ?", id).Find(&products)
		if result.Error != nil {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: result.Error.Error(),
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
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
func SearchProducts(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")

		// Remove spaces from the search query
		query = strings.ReplaceAll(query, " ", "")

		var products []models.Product

		result := db.
			Joins("JOIN category ON products.category_id = category.id").
			Where("REPLACE(LOWER(category.category), ' ', '') LIKE ?", "%"+strings.ToLower(query)+"%").
			Or("REPLACE(LOWER(products.name), ' ', '') LIKE ?", "%"+strings.ToLower(query)+"%").
			Find(&products)

		if result.Error != nil {
			log.Fatal(result.Error)
		}

		jsonData, err := json.Marshal(products)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}
