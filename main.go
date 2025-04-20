package main

import (
	"AccessRefreshToken/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("[ERROR] func: main.init() --> godotenv.Load() | error: %v\n", err)
	}

	if err := handlers.Init(); err != nil {
		log.Fatalf("[ERROR] func: main.init() --> ... --> database.Connect() | error: %v\n", err)
	}
}

func main() {
	Init()

	r := mux.NewRouter()
	r.HandleFunc("/auth/tokens", handlers.GetTokens).Methods(http.MethodGet)
	r.HandleFunc("/auth/refresh", handlers.RefreshTokens).Methods(http.MethodPost)

	port, exists := os.LookupEnv("PORT")

	if !exists {
		log.Fatalln("Dont find PORT in .env")
	}

	http.ListenAndServe(":"+port, r)
}
