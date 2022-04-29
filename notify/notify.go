package notify

import (
	"log"
	"os"
	"strings"

	"github.com/containrrr/shoutrrr"
)

var (
	notificationUrls = os.Getenv("NOTIFICATION_URLS")
)

func Message(text string) {
	urls := strings.Split(notificationUrls, " ")

	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		err := shoutrrr.Send(url, text)
		if err != nil {
			log.Printf("notify: %v", err)
		}
	}
}
