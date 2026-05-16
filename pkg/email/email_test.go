package email

import (
	"os"
	"testing"
)

func TestNewEmailClient(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("MAILTRAP_API_TOKEN", "test-token")
	os.Setenv("MAILTRAP_FROM_EMAIL", "test@example.com")
	os.Setenv("MAILTRAP_FROM_NAME", "Test")

	client := NewEmailClient()

	if client.apiToken != "test-token" {
		t.Errorf("Expected api token 'test-token', got %s", client.apiToken)
	}

	if client.fromEmail != "test@example.com" {
		t.Errorf("Expected from email 'test@example.com', got %s", client.fromEmail)
	}

	if client.fromName != "Test" {
		t.Errorf("Expected from name 'Test', got %s", client.fromName)
	}
}

func TestSendEmailMissingToken(t *testing.T) {
	os.Setenv("MAILTRAP_API_TOKEN", "")
	os.Setenv("MAILTRAP_FROM_EMAIL", "test@example.com")
	os.Setenv("MAILTRAP_FROM_NAME", "Test")

	client := NewEmailClient()
	err := client.SendEmail("recipient@example.com", "Test", "Test body", "", "Test")

	if err == nil {
		t.Error("Expected error when API token is missing")
	}
}
