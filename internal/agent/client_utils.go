package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"unicode/utf8"
)

type Session struct {
	Id string `json::"id"`
}

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (c *openCodeClient) createSession() string {
	slog.Info("Req: /session (create)")
	url := fmt.Sprintf("%s/session", c.baseUrl)
	body := map[string]string{
		"title": fmt.Sprintf("analyze_%d", c.requestCounter),
	}
	encoded_body, err := json.Marshal(body)
	if err != nil {
		log.Fatal("Json encoding failed in session creation")
	}
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(encoded_body))
	if err != nil {
		log.Fatal("Session creation failed")
	}
	var session Session
	error := json.NewDecoder(resp.Body).Decode(&session)
	if error != nil {
		log.Fatal("Json response decoding failed")
	}
	return session.Id
}

func (c *openCodeClient) deleteSession(id string) bool {
	slog.Info("Req: /session (delete)")
	url := fmt.Sprintf("%s/session/%s", c.baseUrl, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	client := c.httpClient
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}
	defer resp.Body.Close()
	var ok bool
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		log.Fatal(err)
	}
	return ok
}

func (c *openCodeClient) sendMessage(id string, input string) {
	slog.Info("Req: /session/:id/message (post)")
	url := fmt.Sprintf("%s/session/%s/message", c.baseUrl, id)
	msg := Message{Type: "text", Text: input}
	encoded_body, err := json.Marshal(map[string][]Message{"parts": []Message{msg}})
	if err != nil {
		log.Fatal("Failed to encode message")
	}
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(encoded_body))
	debugByteArray(resp.Body)
}

func debugByteArray(r io.Reader) {
	data, err := io.ReadAll(r)
	if err != nil {
		log.Fatal("Failed to parse byte array", err)
	}
	if json.Valid(data) {
		fmt.Println("byte array is valid json")
		var out bytes.Buffer
		if err := json.Indent(&out, data, "", "  "); err != nil {
			log.Fatal(err)
		}
		fmt.Println(out.String())
		return
	}
	if utf8.Valid(data) {
		fmt.Println("byte array is utf8 encoded")
		fmt.Println(string(data))
		return
	}
	fmt.Println("byte array is neither json nor utf-8 encoded")
	return
}
