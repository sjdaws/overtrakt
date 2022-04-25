package notify

import (
	"github.com/gregdel/pushover"
	"log"
	"os"
)

var (
	appToken = os.Getenv("PUSHOVER_APP_TOKEN")
	userKey  = os.Getenv("PUSHOVER_USER_KEY")
)

func Message(text string) {
	if appToken == "" || userKey == "" {
		return
	}

	app := pushover.New(appToken)
	recipient := pushover.NewRecipient(userKey)
	message := pushover.NewMessage(text)

	_, err := app.SendMessage(message, recipient)
	if err != nil {
		log.Printf("pushover: %v", err)
	}
}
