package main

import (
	"context"
	"fmt"
	"github.com/aattwwss/ihf-referee-rules/internal"
	"github.com/aattwwss/ihf-referee-rules/trainer"
	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"html/template"
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

	repo := trainer.NewRepository(db)
	service := trainer.NewService(repo)
	//controller := trainer.NewController(service)

	// Define the directory containing your HTML, CSS, and JS files
	dir := "./public"

	// Create a file server to serve static files from the directory
	fileServer := http.FileServer(http.Dir(dir + "/static"))

	// Handle requests to /static/ using the file server
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Handle the root URL ("/") by serving an HTML file (e.g., index.html)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		question, err := service.GetRandomQuestion(ctx)
		if err != nil {
			log.Printf("Error getting random question: %s", err)
		}
		tmpl, err := template.ParseFiles(dir+"/html/base.html", dir+"/html/game.html")
		if err != nil {
			log.Printf("Error parsing template: %s", err)
		}
		err = tmpl.Execute(w, question)
		if err != nil {
			log.Printf("Error executing template: %s", err)
		}
	})

	http.HandleFunc("POST /submit", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		log.Printf("Form submitted: %s", r.Form)
	})

	// Set up and start the HTTP server on port 8080
	port := "8080"
	log.Printf("Server is listening on :%s...", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
