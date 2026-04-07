package telegram_bot

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (b *TelegramBot) SendMessage(m string) error {
	payload := fmt.Sprintf(`{ "chat_id": "%s", "text": "%s"}`, b.chat_id, m)
	res, err := b.httpClient.Post(
		b.url+"/sendMessage",
		"application/json",
		bytes.NewBufferString(payload),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	var tgResponse struct {
		Ok          bool           `json:"ok"`
		Result      map[string]any `json:"result,omitempty"`
		ErrorCode   int            `json:"error_code,omitempty"`
		Description string         `json:"description,omitempty"`
	}
	err = json.NewDecoder(res.Body).Decode(&tgResponse)
	if err != nil {
		return err
	}
	if !tgResponse.Ok {
		return fmt.Errorf("telegram error %d: %s", tgResponse.ErrorCode, tgResponse.Description)
	}

	return nil
}
