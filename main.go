package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"mailer/middleware"
)

type EmailRequest struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

var (
	smtpHost    string
	smtpPort    string
	senderEmail string
	emailAuth   smtp.Auth
)

func init() {
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	senderEmail = os.Getenv("GMAIL_ADDRESS")
	senderPassword := os.Getenv("GMAIL_APP_PASSWORD")

	if smtpHost == "" || smtpPort == "" || senderEmail == "" || senderPassword == "" {
		log.Fatal("Erro: Credenciais SMTP ausentes ou incompletas.")
	}

	emailAuth = smtp.PlainAuth("", senderEmail, senderPassword, smtpHost)
}

func main() {
	http.Handle("/send-email", middleware.CORSMiddleware(middleware.AuthMiddleware(http.HandlerFunc(sendEmailHandler))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Servidor iniciado na porta %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var emailRequest EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailRequest); err != nil {
		http.Error(w, "Erro ao processar o corpo da solicitação", http.StatusBadRequest)
		return
	}

	if emailRequest.To == "" || emailRequest.Subject == "" || emailRequest.Body == "" {
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
		return
	}

	if err := sendEmail(emailRequest); err != nil {
		http.Error(w, "Erro ao enviar o e-mail: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "E-mail enviado com sucesso!")
}

func sendEmail(request EmailRequest) error {
	to := []string{request.To}
	msg := buildEmailMessage(request)

	if err := smtp.SendMail(smtpHost+":"+smtpPort, emailAuth, senderEmail, to, msg); err != nil {
		return fmt.Errorf("erro ao enviar e-mail: %v", err)
	}
	return nil
}

func buildEmailMessage(request EmailRequest) []byte {
	return []byte("MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"To: " + request.To + "\r\n" +
		"Subject: " + request.Subject + "\r\n" +
		"\r\n" +
		request.Body + "\r\n")
}
