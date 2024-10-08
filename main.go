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
	err := json.NewDecoder(r.Body).Decode(&emailRequest)
	if err != nil {
		http.Error(w, "Erro ao processar o corpo da solicitação", http.StatusBadRequest)
		return
	}

	if emailRequest.To == "" || emailRequest.Subject == "" || emailRequest.Body == "" {
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
		return
	}

	err = sendEmail(emailRequest)
	if err != nil {
		http.Error(w, "Erro ao enviar o e-mail: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "E-mail enviado com sucesso!")
}

func sendEmail(request EmailRequest) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	sender := os.Getenv("GMAIL_ADDRESS")
	password := os.Getenv("GMAIL_APP_PASSWORD")

	if sender == "" || password == "" {
		return fmt.Errorf("Credenciais do SMTP não definidas")
	}

	auth := smtp.PlainAuth("", sender, password, smtpHost)

	to := []string{request.To}

	msg := []byte("MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"To: " + request.To + "\r\n" +
		"Subject: " + request.Subject + "\r\n" +
		"\r\n" +
		request.Body + "\r\n")

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, sender, to, msg)
	if err != nil {
		return fmt.Errorf("erro ao enviar e-mail: %v", err)
	}

	return nil
}
