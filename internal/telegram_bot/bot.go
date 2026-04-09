package telegram_bot

import (
	"bytes"
	"diane/internal/agent"
	"encoding/json"
	"fmt"
	"net/http"
)

type NotificationService interface {
	SendMessage(agent.Notification) error
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

func (b *TelegramBot) SendMessage(message agent.Notification) error {
	text := message.ToFormattedText()
	if text != "" {
		if err := b.post("/sendMessage", map[string]any{
			"chat_id": b.chat_id,
			"text":    text,
		}); err != nil {
			return err
		}
	}

	photos := message.ToPhotos()
	switch len(photos) {
	case 0:
		return nil
	case 1:
		return b.post("/sendPhoto", map[string]any{
			"chat_id": b.chat_id,
			"photo":   photos[0],
		})
	default:
		media := make([]map[string]string, 0, len(photos))
		for _, photo := range photos {
			media = append(media, map[string]string{
				"type":  "photo",
				"media": photo,
			})
		}
		return b.post("/sendMediaGroup", map[string]any{
			"chat_id": b.chat_id,
			"media":   media,
		})
	}
}

func (b *TelegramBot) post(path string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := b.httpClient.Post(
		b.url+path,
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return err
	}
	defer res.Body.Close()

	var tgResponse struct {
		Ok          bool            `json:"ok"`
		Result      json.RawMessage `json:"result,omitempty"`
		ErrorCode   int             `json:"error_code,omitempty"`
		Description string          `json:"description,omitempty"`
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
