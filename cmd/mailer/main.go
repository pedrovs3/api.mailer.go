package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"mailer/internal/email"
	"mailer/internal/middleware"
)

func main() {
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
