package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/mail"
	"os"
	"sync"
	"time"
	"mailer/middleware"
)

type EmailRequest struct {
	From    string `json:"from,omitempty"` 
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

var (
	smtpHost      string
	smtpPort      string
	senderEmail   string
	emailAuth     smtp.Auth
	authOnce      sync.Once
)

func init() {
	authOnce.Do(initSMTPAuth)
}

func initSMTPAuth() {
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

	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // Limitar a leitura para 1MB
	defer r.Body.Close()

	var emailRequest EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailRequest); err != nil {
		http.Error(w, "Erro ao processar o corpo da solicitação", http.StatusBadRequest)
		return
	}

	if !validateEmail(emailRequest.To) {
		http.Error(w, "Endereço de e-mail inválido", http.StatusBadRequest)
		return
	}

	if emailRequest.To == "" || emailRequest.Subject == "" || emailRequest.Body == "" {
		http.Error(w, "Todos os campos são obrigatórios", http.StatusBadRequest)
		return
	}

	go func() {
		if err := sendEmail(emailRequest); err != nil {
			log.Printf("Erro ao enviar e-mail: %v", err)
		} else {
			log.Printf("E-mail enviado com sucesso para %s", emailRequest.To)
		}
	}()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "E-mail enviado com sucesso!")
}

func sendEmail(request EmailRequest) error {
	to := []string{request.To}

	from := senderEmail
	if request.From != "" {
		if !validateEmail(request.From) {
			return fmt.Errorf("remetente inválido")
		}
		from = request.From
	}

	msg := buildEmailMessage(from, request)

	for retries := 0; retries < 3; retries++ {
		if err := smtp.SendMail(smtpHost+":"+smtpPort, emailAuth, from, to, msg); err != nil {
			log.Printf("Tentativa %d de envio de e-mail falhou: %v", retries+1, err)
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}

	return fmt.Errorf("falha ao enviar e-mail após várias tentativas")
}

func buildEmailMessage(from string, request EmailRequest) []byte {
	return []byte("MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"From: " + from + "\r\n" +
		"To: " + request.To + "\r\n" +
		"Subject: " + request.Subject + "\r\n" +
		"\r\n" +
		request.Body + "\r\n")
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
