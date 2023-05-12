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

	"git.tecblic.com/sanyog-tecblic/ecom/controller/models"
	"github.com/go-chi/chi"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var newID int

func CreateItemHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// err := os.Mkdir("uploads", 0755)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		w.Header().Set("Content-Type", "application/json")
		r.ParseMultipartForm(10 << 20)

		category := r.FormValue("category")

		image, handler, err := r.FormFile("image")

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
		if category == "" {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: "Category name is required",
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		stmt, err := db.Prepare(`INSERT INTO category (category,imageurl) VALUES ($1,$2)`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w, "Error preparing SQL statement: %v", err)
			return
		}
		defer stmt.Close()

		result, err := stmt.Exec(category, imageURL)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error executing SQL statement: %v", err)
			return
		}

		id, _ := result.LastInsertId()

		categories := models.Category{
			ID:         int(id),
			Category:   category,
			Statuscode: http.StatusCreated,
			ImageURL:   imageURL,
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Category created successfully")
		json.NewEncoder(w).Encode(categories)
	}
}

func GetAllCategpries(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var categories []models.Category
		// var row_var error

		rows, err := db.Query("SELECT id,category,imageurl FROM category")

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var category models.Category
			row_var := rows.Scan(&category.ID, &category.Category, &category.ImageURL)
			if row_var != nil {
				log.Fatal(row_var)
			}
			categories = append(categories, category)
		}
		json.NewEncoder(w).Encode(categories)
	}
}

func UpdateCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		CategoryID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Println(err)
		}
		var category models.Category

		json.NewDecoder(r.Body).Decode(&category)
		if err != nil {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: "Error: " + err.Error(),
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		if category.Category == "" {
			apierror := models.APIError{
				Code:    http.StatusBadRequest,
				Message: "Category is required",
			}
			w.WriteHeader(apierror.Code)
			json.NewEncoder(w).Encode(apierror)
			return
		}
		result, err := db.Exec("update category set category=$1 where id=$2", category.Category, CategoryID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error updating category: %s", err.Error())
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
			fmt.Fprintf(w, "Category not found with ID: %s", CategoryID)
			return
		}
		if err == nil {
			// id := strconv.Itoa(TaskID)
			category = models.Category{
				ID:         CategoryID,
				Category:   category.Category,
				Statuscode: http.StatusOK,
			}
			// response := map[string]string{"id": id, "message": "User updated successfully", "tasks": task.Tasks, "statuscode": http.StatusText(http.StatusOK)}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(http.StatusOK)
			json.NewEncoder(w).Encode(category)
		}
	}
}
func Deletecategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		param := mux.Vars(r)
		CategoryID, err := strconv.Atoi(param["id"])
		// CategoryID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Println(err)
		}
		// var task model.Task
		// query := fmt.Sprintf("delete from todo where id=$1", TaskID)
		result, err := db.Exec("delete from category where id=$1", CategoryID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error deleting Category: %s", err.Error())
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
			fmt.Fprintf(w, "Category not found with ID: %s", CategoryID)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Category deleted successfully!")
	}
}
