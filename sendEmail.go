package sendemail

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	_ "github.com/lib/pq"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/rs/zerolog"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type UserData struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func SendVerificationEmail(ctx context.Context, m pubsub.Message) error {
	// Initialize zerolog logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	var userData UserData
	pubSubMsg := string(m.Data)

	fmt.Println("Data: ", pubSubMsg)

	// base64Data, err := base64.URLEncoding.DecodeString(pubSubMsg)

	// fmt.Println("Base64Data", string(base64Data))

	// if err != nil {
	// 	fmt.Println("Error decoding string:", err)
	// 	return err
	// }

	if err := json.Unmarshal([]byte(pubSubMsg), &userData); err != nil {
		logger.Error().Err(err).Msg("error decoding data")
		return err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_HOST"), 5432, os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBNAME"))

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to connect to database")
		return err
	}
	defer db.Close()

	// Initialize Mailgun client
	domain := os.Getenv("SENDER_DOMAIN")
	apiKey := os.Getenv("MAILGUN_API_KEY")
	mg := mailgun.NewMailgun(domain, apiKey)
	sender := "no-reply@babuaravind-gururaj.me"
	subject := "Verify Your Email Address"
	verifyLink := fmt.Sprintf("http://%s:8080/v1/user/verify?token=%s", domain, userData.Token)
	body := fmt.Sprintf("Please verify your email by clicking on the link: %s", verifyLink)
	recipient := userData.Email

	fmt.Println("Verify link: ", verifyLink)
	fmt.Println("Recipient: ", recipient)

	// Create the message
	message := mg.NewMessage(sender, subject, body, recipient)

	// Send the message
	resp, id, err := mg.Send(ctx, message)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send email")
		return err
	}

	// Update email_trigger_time before sending email
	if err := updateEmailTriggerTime(ctx, db, userData.Token); err != nil {
		logger.Error().Err(err).Msg("Failed to update email_trigger_time")
		return err
	}

	logger.Info().Str("id", id).Str("response", resp).Msg("Sent email and updated email_trigger_time")
	return nil
}

func updateEmailTriggerTime(ctx context.Context, db *sql.DB, userID string) error {
	// SQL query to update email_trigger_time
	_, err := db.ExecContext(ctx, "UPDATE verify_users SET email_trigger_time = now() WHERE token = $1", userID)
	if err != nil {
		return err
	}
	return nil
}
