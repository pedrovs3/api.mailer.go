package config

import (
	"log"
	"os"
)

func LoadEnvVariables() {
	requiredVars := []string{"SMTP_HOST", "SMTP_PORT", "GMAIL_ADDRESS", "GMAIL_APP_PASSWORD", "API_KEY", "PORT"}
	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("A variável de ambiente %s não foi definida", envVar)
		}
	}
}
