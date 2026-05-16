package email


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type EmailClient struct {
	apiToken string
	apiURL   string
	fromEmail string
	fromName string
}

type EmailPayload struct {
	From struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"from"`
	To []struct {
		Email string `json:"email"`
	} `json:"to"`
	Subject  string `json:"subject"`
	Text     string `json:"text"`
	Html     string `json:"html,omitempty"`
	Category string `json:"category"`
}

func NewEmailClient() *EmailClient {
	return &EmailClient{
		apiToken: os.Getenv("MAILTRAP_API_TOKEN"),
		apiURL:   "https://send.api.mailtrap.io/api/send",
		fromEmail: os.Getenv("MAILTRAP_FROM_EMAIL"),
		fromName: os.Getenv("MAILTRAP_FROM_NAME"),
	}
}

func (e *EmailClient) SendEmail(toEmail string, subject string, textBody string, htmlBody string, category string) error {
	if e.apiToken == "" {
		return fmt.Errorf("MAILTRAP_API_TOKEN not configured")
	}

	payload := EmailPayload{
		Subject:  subject,
		Text:     textBody,
		Html:     htmlBody,
		Category: category,
	}

	payload.From.Email = e.fromEmail
	payload.From.Name = e.fromName
	payload.To = append(payload.To, struct {
		Email string `json:"email"`
	}{Email: toEmail})

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", e.apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", e.apiToken))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Mailtrap Response [%d]: %s\n", res.StatusCode, string(body))

	if res.StatusCode >= 400 {
		return fmt.Errorf("mailtrap error: status %d, body: %s", res.StatusCode, string(body))
	}

	return nil
}

func (e *EmailClient) SendWelcomeEmail(toEmail string, username string) error {
	subject := "Selamat datang di CWD Forum!"
	text := fmt.Sprintf("Halo %s,\n\nTerima kasih telah mendaftar di CWD Forum. Kami senang memiliki Anda di komunitas kami.", username)
	html := fmt.Sprintf(`<h1>Selamat datang %s!</h1><p>Terima kasih telah mendaftar di CWD Forum.</p>`, username)

	return e.SendEmail(toEmail, subject, text, html, "Registration")
}

func (e *EmailClient) SendForgotPasswordEmail(toEmail string, resetLink string) error {
	subject := "Reset Password - CWD Forum"
	text := fmt.Sprintf("Klik link berikut untuk reset password: %s", resetLink)
	html := fmt.Sprintf(`<h1>Reset Password</h1><p>Klik <a href="%s">di sini</a> untuk reset password Anda.</p>`, resetLink)

	return e.SendEmail(toEmail, subject, text, html, "ForgotPassword")
}

func (e *EmailClient) SendNotificationEmail(toEmail string, notificationType string, message string) error {
	subject := fmt.Sprintf("Notifikasi - %s", notificationType)
	text := message
	html := fmt.Sprintf(`<p>%s</p>`, message)

	return e.SendEmail(toEmail, subject, text, html, "Notification")
}
