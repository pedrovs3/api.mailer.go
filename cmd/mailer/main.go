package main

import (
	"fmt"
	"log"
	"mailer/internal/email"
	"mailer/internal/middleware"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "API is healthy")
	})

	http.Handle("/send-email",
		middleware.CORSMiddleware(
			middleware.AuthMiddleware(
				middleware.RateLimitMiddleware(
					http.HandlerFunc(email.EnqueueEmailHandler),
				),
			),
		),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Servidor iniciado na porta %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
