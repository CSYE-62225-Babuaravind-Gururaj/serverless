package sendemail

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/rs/zerolog"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type UserData struct {
	Email      string `json:"email"`
	VerifyLink string `json:"verifyLink"`
	UserID     string `json:"userID"`
}

func updateVerificationStatus(ctx context.Context, userID string) error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		// handle error
	}
	defer db.Close()

	// Example SQL query to update verification status
	_, err = db.ExecContext(ctx, "UPDATE verify_users SET email_verified = TRUE WHERE user_id = $1", userID)
	if err != nil {
		// handle error
	}
	return nil
}

func SendVerificationEmail(ctx context.Context, m pubsub.Message) error {
	// Initialize zerolog logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	var userData UserData
	if err := json.Unmarshal(m.Data, &userData); err != nil {
		logger.Error().Err(err).Msg("error decoding data")
		return err
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return err
	}
	defer db.Close()

	// Update email_trigger_time before sending email
	if err := updateEmailTriggerTime(ctx, db, userData.UserID); err != nil {
		logger.Error().Err(err).Msg("Failed to update email_trigger_time")
		return err
	}

	// Initialize Mailgun client
	domain := "babuaravind-gururaj.me"
	apiKey := "68cce657fa6d365a396c07445aaf856f-f68a26c9-fb58f577"
	mg := mailgun.NewMailgun(domain, apiKey)
	sender := "no-reply@babuaravind-gururaj.me"
	subject := "Verify Your Email Address"
	body := fmt.Sprintf("Please verify your email by clicking on the link: %s", userData.VerifyLink)
	recipient := userData.Email

	// Create the message
	message := mg.NewMessage(sender, subject, body, recipient)

	// Send the message
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send email")
		return err
	}

	logger.Info().Str("id", id).Str("response", resp).Msg("Sent email and updated email_trigger_time")
	return nil
}

func updateEmailTriggerTime(ctx context.Context, db *sql.DB, userID string) error {
	// SQL query to update email_trigger_time
	_, err := db.ExecContext(ctx, "UPDATE verify_users SET email_trigger_time = now() WHERE id = $1", userID)
	if err != nil {
		return err
	}
	return nil
}
