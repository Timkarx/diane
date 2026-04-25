package core

import (
	"fmt"
	"strings"
)

type Notification interface {
	ToFormattedText() string
	ToPhotos() []string
}

type ListingNotification struct {
	Text   string
	Link   string
	Photos []string
}

func (n ListingNotification) ToFormattedText() string {
	text := strings.TrimSpace(n.Text)
	link := strings.TrimSpace(n.Link)

	switch {
	case text == "":
		return link
	case link == "":
		return text
	default:
		return fmt.Sprintf("%s\n\n%s", text, link)
	}
}

func (n ListingNotification) ToPhotos() []string {
	cleaned := make([]string, 0, len(n.Photos))
	for _, photo := range n.Photos {
		photo = strings.TrimSpace(photo)
		if photo == "" {
			continue
		}
		cleaned = append(cleaned, photo)
	}
	return cleaned
}


