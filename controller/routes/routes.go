package routes

import (
	"log"
	"net/http"

	"git.tecblic.com/sanyog-tecblic/ecom/controller/endpoints"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Routes() http.Handler {
	dsn := "host=localhost user=postgres password=root dbname=ecom port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/login", endpoints.LoginHandler(db)).Methods("POST")

	r.HandleFunc("/category", endpoints.GetAllCategories(db)).Methods("GET")

	r.Handle("/product/{id}", endpoints.GetAllProductsByCID(db)).Methods("GET")
	r.Handle("/product", endpoints.GetAllProducts(db)).Methods("GET")

	r.Handle("/Register", endpoints.Register(db)).Methods("POST")

	r.HandleFunc("/product", endpoints.CreateProduct(db)).Methods("POST")

	r.Handle("/search", endpoints.SearchProducts(db)).Methods("GET")

	r.Handle("/updateprofile", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.UpdateProfile(db)))).Methods("PATCH")

	r.Handle("/getcart", endpoints.AuthMiddleware(endpoints.CartItems(db))).Methods("GET")
	r.Handle("/products/{id}", endpoints.ViewCartNotLoggedIn(db)).Methods("GET")
	r.Handle("/cart", endpoints.DeleteCart(db)).Methods("DELETE")

	r.Handle("/profile", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.GetUserProfile(db)))).Methods("GET")

	r.Handle("/cart", endpoints.DeleteCart(db)).Methods("DELETE")

	r.Handle("/updatequantityprice", endpoints.AuthMiddleware(http.HandlerFunc(endpoints.UpdateQuantityAndPrice(db)))).Methods("PATCH")

	r.Handle("/cart", endpoints.AddToCart(db)).Methods("POST")
	r.Handle("/cart2", endpoints.AddToCart2(db)).Methods("POST")

	// ... other route definitions ...

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
