package main

import (
	"flag"
	"fmt"
	"gin-quickstart/pkg/email"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Define flags
	register := flag.Bool("register", false, "Test welcome email (registration)")
	forgotPassword := flag.Bool("forgot-password", false, "Test forgot password email")
	notification := flag.Bool("notification", false, "Test notification email")
	toEmail := flag.String("email", "owner@example.com", "Recipient email address")
	flag.Parse()

	fmt.Println("=== Mailtrap Configuration ===")
	fmt.Printf("API Token: %s\n", os.Getenv("MAILTRAP_API_TOKEN"))
	fmt.Printf("From Email: %s\n", os.Getenv("MAILTRAP_FROM_EMAIL"))
	fmt.Printf("From Name: %s\n", os.Getenv("MAILTRAP_FROM_NAME"))
	fmt.Println()

	client := email.NewEmailClient()

	var err error

	switch {
	case *register:
		fmt.Println("📧 Sending welcome email (registration)...")
		err = client.SendWelcomeEmail(*toEmail, "TestUser")
	case *forgotPassword:
		fmt.Println("🔑 Sending forgot password email...")
		resetLink := "https://cwdforum.com/reset-password?token=test-token-123"
		err = client.SendForgotPasswordEmail(*toEmail, resetLink)
	case *notification:
		fmt.Println("🔔 Sending notification email...")
		err = client.SendNotificationEmail(*toEmail, "Thread Reply", "Ada balasan baru pada thread Anda")
	default:
		fmt.Println("Usage:")
		fmt.Println("  go run test_email_send.go -register -email=owner@example.com")
		fmt.Println("  go run test_email_send.go -forgot-password -email=owner@example.com")
		fmt.Println("  go run test_email_send.go -notification -email=owner@example.com")
		fmt.Println("\n💡 Tip: Ganti 'owner@example.com' dengan email yang Anda gunakan untuk login ke Mailtrap")
		return
	}

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("\n💡 Tip: Pastikan email adalah email yang Anda gunakan untuk login ke Mailtrap")
		return
	}

	fmt.Println("✅ Email sent successfully!")
	fmt.Println("\n📧 Check your Mailtrap inbox at: https://mailtrap.io/inboxes")
}
