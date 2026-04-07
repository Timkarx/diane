package telegram_bot

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type NotificationService interface {
	SendMessage(string) (bool, error)
}

type TelegramBot struct {
	httpClient *http.Client
	url        string
	chat_id    string
}

func NewTelegramBot(token string, chat_id string) TelegramBot {
	api_url := "https://api.telegram.org/bot" + token
	bot := TelegramBot{
		url:        api_url,
		httpClient: http.DefaultClient,
		chat_id:    chat_id,
	}

	return bot
}

func (b *TelegramBot) SendMessage(m string) (bool, error) {
	payload := fmt.Sprintf(`{ "chat_id": "%s", "text": "%s"}`, b.chat_id, m)
	fmt.Println(payload)
	res, err := b.httpClient.Post(
		b.url+"/sendMessage",
		"application/json",
		bytes.NewBufferString(payload),
	)

	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	fmt.Printf("telegram response status: %s\n", res.Status)
	fmt.Printf("telegram response body: %s\n", string(body))

	return true, nil
}
