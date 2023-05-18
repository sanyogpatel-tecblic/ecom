package routes

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/config"
	"git.tecblic.com/sanyog-tecblic/ecom/controller/endpoints"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func Routes(app *config.AppConfig) http.Handler {
	db, err := sql.Open("postgres", "postgresql://postgres:root@localhost/ecom?sslmode=disable")
	fmt.Println("connected to database!")
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/category", endpoints.CreateItemHandler(db)).Methods("POST")
	r.HandleFunc("/category", endpoints.GetAllCategpries(db)).Methods("GET")
	r.HandleFunc("/category/{id}", endpoints.UpdateCategory(db)).Methods("PUT")
	r.HandleFunc("/category/{id}", endpoints.Deletecategory(db)).Methods("DELETE")

	r.HandleFunc("/product", endpoints.CreateProduct(db)).Methods("POST")
	r.Handle("/product", endpoints.GetAllProducts(db)).Methods("GET")
	r.Handle("/product/{id}", endpoints.GetAllProductsByCID(db)).Methods("GET")

	r.HandleFunc("/login", endpoints.LoginHandler(db)).Methods("POST")
	r.Handle("/Register", endpoints.Register(db)).Methods("POST")

	r.Handle("/updateprofile", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.UpdateProfile(db)))).Methods("PATCH")
	r.Handle("/updateprofilepicture", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.EditProfilePicture(db)))).Methods("PATCH")
	r.Handle("/updatequantityprice", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.UpdatequantityAndPrice(db)))).Methods("PATCH")
	// r.Handle("/updatepersonalinfo", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.UpdatePersonalInfo(db)))).Methods("PATCH")

	r.Handle("/getusers", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.GetAllUsers(db)))).Methods("GET")
	r.Handle("/cart", endpoints.AddToCart(db)).Methods("POST")
	r.Handle("/cart", endpoints.DeleteCart(db)).Methods("DELETE")
	r.Handle("/products/{id}", endpoints.ViewCartNotLoggedIn(db)).Methods("GET")

	r.Handle("/getcart", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.CartItems(db)))).Methods("GET")
	// r.Handle("/cart", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.GetCartHandler(db)))).Methods("GET")
	// r.Handle("/profile", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.Userprofile(db)))).Methods("GET")
	r.Handle("/profile", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.GetUserProfile(db)))).Methods("GET")
	// r.Handle("/profile", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.UpdateProfile(db)))).Methods("PUT")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost", "http://192.168.0.39:5500"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := c.Handler(r)
	log.Fatal(http.ListenAndServe(":8050", handler))

	// Return the cors middleware handler instead of the mux router handler
	return handler
}
