package main

import (
	"context"
	"fmt"
	"github.com/aattwwss/ihf-referee-rules/internal"
	"github.com/aattwwss/ihf-referee-rules/public"
	"github.com/aattwwss/ihf-referee-rules/trainer"
	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

func main() {

	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := internal.EnvConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	connectionUrl := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", cfg.DbUsername, cfg.DbPassword, cfg.DbHost, cfg.DbPort, cfg.DbDatabase)
	db, err := pgxpool.New(ctx, connectionUrl)
	if err != nil {
		log.Fatal(err)
	}

	htmlFS, err := public.HTML()
	if err != nil {
		log.Fatal(err)
	}

	staticFS, err := public.Static()
	if err != nil {
		log.Fatal(err)
	}

	repo := trainer.NewRepository(db)
	service := trainer.NewService(repo)
	controller := trainer.NewController(service, htmlFS)

	// Create a file server to serve static files from the directory
	fileServer := http.FileServer(http.FS(staticFS))

	// Handle requests to /static/ using the file server
	http.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	http.HandleFunc("GET /health", controller.Health)

	http.HandleFunc("GET /", controller.Home)

	http.HandleFunc("GET /feedback", controller.Feedback)
	http.HandleFunc("POST /feedback", controller.SubmitFeedback)

	http.HandleFunc("GET /quiz", controller.QuizConfig)
	http.HandleFunc("POST /quiz", controller.SubmitQuizConfig)
	http.HandleFunc("GET /quiz/{key}", controller.DoQuiz)
	http.HandleFunc("POST /quiz/{key}", controller.SubmitQuiz)

	// Set up and start the HTTP server on port 8080
	port := "8080"
	log.Printf("Server is listening on :%s...", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
