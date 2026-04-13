package telegram_bot

import (
	"diane/internal/core"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessageSendsTextWithLink(t *testing.T) {
	type request struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}

	var gotPath string
	var gotBody request

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	bot := TelegramBot{
		httpClient: server.Client(),
		url:        server.URL,
		chat_id:    "chat-123",
	}

	err := bot.SendMessage(core.ListingNotification{
		Text: "Fits the criteria",
		Link: "https://example.com/listing",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if gotPath != "/sendMessage" {
		t.Fatalf("path = %q, want %q", gotPath, "/sendMessage")
	}
	if gotBody.ChatID != "chat-123" {
		t.Fatalf("chat_id = %q, want %q", gotBody.ChatID, "chat-123")
	}
	if gotBody.Text != "Fits the criteria\n\nhttps://example.com/listing" {
		t.Fatalf("text = %q, want %q", gotBody.Text, "Fits the criteria\n\nhttps://example.com/listing")
	}
}

func TestSendMessageSendsPhotoGroupWhenPhotosPresent(t *testing.T) {
	type textRequest struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	type mediaItem struct {
		Type  string `json:"type"`
		Media string `json:"media"`
	}
	type mediaRequest struct {
		ChatID string      `json:"chat_id"`
		Media  []mediaItem `json:"media"`
	}

	var paths []string
	var gotText textRequest
	var gotMedia mediaRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)

		switch r.URL.Path {
		case "/sendMessage":
			if err := json.NewDecoder(r.Body).Decode(&gotText); err != nil {
				t.Fatalf("decode text request body: %v", err)
			}
			_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
		case "/sendMediaGroup":
			if err := json.NewDecoder(r.Body).Decode(&gotMedia); err != nil {
				t.Fatalf("decode media request body: %v", err)
			}
			_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	bot := TelegramBot{
		httpClient: server.Client(),
		url:        server.URL,
		chat_id:    "chat-123",
	}

	err := bot.SendMessage(core.ListingNotification{
		Link: "https://example.com/listing",
		Photos: []string{
			"https://example.com/1.jpg",
			"https://example.com/2.jpg",
		},
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("len(paths) = %d, want %d", len(paths), 2)
	}
	if paths[0] != "/sendMessage" {
		t.Fatalf("paths[0] = %q, want %q", paths[0], "/sendMessage")
	}
	if paths[1] != "/sendMediaGroup" {
		t.Fatalf("paths[1] = %q, want %q", paths[1], "/sendMediaGroup")
	}
	if gotText.ChatID != "chat-123" {
		t.Fatalf("text chat_id = %q, want %q", gotText.ChatID, "chat-123")
	}
	if gotText.Text != "https://example.com/listing" {
		t.Fatalf("text = %q, want %q", gotText.Text, "https://example.com/listing")
	}
	if gotMedia.ChatID != "chat-123" {
		t.Fatalf("media chat_id = %q, want %q", gotMedia.ChatID, "chat-123")
	}
	if len(gotMedia.Media) != 2 {
		t.Fatalf("len(media) = %d, want %d", len(gotMedia.Media), 2)
	}
	if gotMedia.Media[0].Media != "https://example.com/1.jpg" {
		t.Fatalf("media[0] = %q, want %q", gotMedia.Media[0].Media, "https://example.com/1.jpg")
	}
	if gotMedia.Media[1].Media != "https://example.com/2.jpg" {
		t.Fatalf("media[1] = %q, want %q", gotMedia.Media[1].Media, "https://example.com/2.jpg")
	}
}
