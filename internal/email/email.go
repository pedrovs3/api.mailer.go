package email

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"time"
)

type EmailRequest struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func smtpAuth(username, password, host string) smtp.Auth {
	return smtp.PlainAuth("", username, password, host)
}

var (
	smtpConfigs = []SMTPConfig{
		{Host: os.Getenv("SMTP_HOST"), Port: os.Getenv("SMTP_PORT"), Auth: smtpAuth(os.Getenv("GMAIL_ADDRESS"), os.Getenv("GMAIL_APP_PASSWORD"), os.Getenv("SMTP_HOST"))},
	}
	emailQueue  = make(chan EmailRequest, 100)
	rateLimiter = time.Tick(time.Second * 1)
	senderEmail = os.Getenv("GMAIL_ADDRESS")
)

type SMTPConfig struct {
	Host string
	Port string
	Auth smtp.Auth
}

func init() {
	go startWorker()
}

func EnqueueEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var emailRequest EmailRequest
	if err := json.NewDecoder(r.Body).Decode(&emailRequest); err != nil {
		http.Error(w, "Erro ao processar o corpo da solicitação", http.StatusBadRequest)
		return
	}

	if !validateEmail(emailRequest.To) {
		http.Error(w, "Endereço de e-mail inválido", http.StatusBadRequest)
		return
	}

	emailQueue <- emailRequest
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "E-mail adicionado à fila para envio.")
}

func startWorker() {
	for req := range emailQueue {
		select {
		case <-rateLimiter:
			if err := sendWithRetry(req); err != nil {
				log.Printf("Falha ao enviar e-mail após tentativas: %v", err)
			}
		}
	}
}

func sendWithRetry(request EmailRequest) error {
	for _, config := range smtpConfigs {
		if err := sendEmail(request, config); err == nil {
			return nil
		}
		log.Printf("Erro com provedor %s:%s, tentando próximo...", config.Host, config.Port)
	}
	return fmt.Errorf("todos os provedores falharam")
}

func sendEmail(request EmailRequest, config SMTPConfig) error {
	from := request.From
	if from == "" {
		from = senderEmail
	}

	msg := buildEmailMessage(from, request)
	to := []string{request.To}

	for retries := 0; retries < 3; retries++ {
		if err := smtp.SendMail(config.Host+":"+config.Port, config.Auth, from, to, msg); err != nil {
			log.Printf("Tentativa %d de envio de e-mail falhou: %v", retries+1, err)
			time.Sleep(2 * time.Second)
		} else {
			log.Printf("E-mail enviado com sucesso para %s", request.To)
			return nil
		}
	}
	return fmt.Errorf("falha ao enviar e-mail após tentativas")
}

func buildEmailMessage(from string, request EmailRequest) []byte {
	return []byte("MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"From: " + from + "\r\n" +
		"To: " + request.To + "\r\n" +
		"Subject: " + request.Subject + "\r\n" +
		"\r\n" +
		request.Body)
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
