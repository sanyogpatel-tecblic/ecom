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

	"git.tecblic.com/sanyog-tecblic/ecom/controller/models"
)

func GetUserProfile(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		accessToken := r.Header.Get("Authorization")

		if accessToken == "" {
			http.Error(w, "missing access token", http.StatusUnauthorized)
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

		var users []models.User

		rows, err := db.Query(`select id,username,password,email,
								name,gender,mobile,image_url,address from users
								where id=$1`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for rows.Next() {
			var user models.User
			err := rows.Scan(
				&user.ID,
				&user.Username,
				&user.Password,
				&user.Email,
				&user.Name,
				&user.Gender,
				&user.Mobile,
				&user.ImageURL,
				&user.Address)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			users = append(users, user)
		}
		json.NewEncoder(w).Encode(users)
	}
}

// personal infor add endpoint
func UpdateProfile(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "multipart/form-data")

		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "missing access token", http.StatusUnauthorized)
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

		err = r.ParseMultipartForm(32 << 20) // max memory 32MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		username := r.FormValue("username")
		email := r.FormValue("email")
		name := r.FormValue("name")
		gender := r.FormValue("gender")
		mobile := r.FormValue("mobile")
		address := r.FormValue("address")

		var imageURL string

		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			allowedExtensions := map[string]bool{
				".jpg":  true,
				".jpeg": true,
				".png":  true,
			}

			ext := filepath.Ext(handler.Filename)
			if !allowedExtensions[ext] {
				http.Error(w, "Invalid file format. Only JPG, JPEG, PNG files are allowed.", http.StatusBadRequest)
				return
			}

			tempFile, err := os.CreateTemp("", "upload-*"+ext)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error creating temporary file: %v", err), http.StatusInternalServerError)
				return
			}
			defer tempFile.Close()

			io.Copy(tempFile, file)

			imageURL = fmt.Sprintf("../uploads/%s", handler.Filename)
			err = os.Rename(tempFile.Name(), imageURL)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error moving file to uploads directory: %v", err), http.StatusInternalServerError)
				return
			}

		} else if err == http.ErrMissingFile {
			// no image uploaded, keep the existing image URL
			imageURL = r.FormValue("imageurl")
		} else {
			http.Error(w, fmt.Sprintf("Error uploading image: %v", err), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(`UPDATE users SET username = $1, email =$2, name = $3, gender = $4, mobile =$5, address = $6,
			image_url = CASE 
				WHEN $7 <> '' THEN $7 -- new image URL provided
				ELSE image_url -- no image URL provided, keep existing value
			END
			WHERE id = $8;`, username, email, name, gender, mobile, address, imageURL, id)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error preparing SQL statement: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode("Updated")
	}
}

// profile pic update endpoint
func EditProfilePicture(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "missing access token", http.StatusUnauthorized)
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
		err = r.ParseMultipartForm(32 << 20) // max memory 32MB
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

		stmt, err := db.Prepare(`UPDATE users SET image_url=$1 where id=$2`)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(w, "Error preparing SQL statement: %v", err)
			return
		}
		defer stmt.Close()
		_, err = stmt.Exec(imageURL, id)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error executing SQL statement: %v", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Profile picture updated successfully")
		json.NewEncoder(w).Encode("updated")
	}
}

// overall profile update endpoint
func AddPersonalInfo(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		var user models.User
		accessToken := r.Header.Get("Authorization")

		if accessToken == "" {
			http.Error(w, "missing access token", http.StatusUnauthorized)
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
		json.NewDecoder(r.Body).Decode(&user)
		result, err := db.Exec(`update users set address=$1,name=$2,gender=$3,mobile=$4 where id=$5`, user.Address, user.Name, user.Gender, user.Mobile, id)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
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
		json.NewEncoder(w).Encode(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}
