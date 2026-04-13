package opencode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"diane/core"
)

func marshalUnionValue(dst *json.RawMessage, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}

	*dst = encoded
	return nil
}

func (t SessionPromptJSONBody_Parts_Item) MarshalJSON() ([]byte, error) {
	return t.union.MarshalJSON()
}

func (t *SessionPromptJSONBody_Parts_Item) UnmarshalJSON(b []byte) error {
	return t.union.UnmarshalJSON(b)
}

func (t *SessionPromptJSONBody_Parts_Item) FromTextPartInput(v TextPartInput) error {
	return marshalUnionValue(&t.union, v)
}

func newTextPromptPart(input string) (SessionPromptJSONBody_Parts_Item, error) {
	textPart := TextPartInput{
		Text: input,
		Type: "text",
	}

	var part SessionPromptJSONBody_Parts_Item
	if err := part.FromTextPartInput(textPart); err != nil {
		return SessionPromptJSONBody_Parts_Item{}, fmt.Errorf("encode text part: %w", err)
	}

	return part, nil
}

func newOutputFormat[T core.Actionable]() (*OutputFormat, error) {
	var action T
	schema := action.Schema()
	if len(schema) == 0 {
		return nil, nil
	}

	var outputFormat OutputFormat
	err := outputFormat.FromOutputFormatJsonSchema(OutputFormatJsonSchema{
		Schema: schema,
		Type:   "json_schema",
	})
	if err != nil {
		return nil, fmt.Errorf("encode response format: %w", err)
	}

	return &outputFormat, nil
}

func (c *OpencodeAgent[T]) createSession() (Session, error) {
	slog.Info("req /session", "action", "create")

	title := fmt.Sprintf("analyze_%d", c.requestCounter)
	body := SessionCreateJSONBody{Title: &title}

	var session Session
	if err := c.doJSON(http.MethodPost, "/session", body, &session); err != nil {
		return Session{}, err
	}

	c.requestCounter++
	return session, nil
}

func (c *OpencodeAgent[T]) deleteSession(id string) (bool, error) {
	slog.Info("req /session/{sessionID}", "action", "delete", "session_id", id)

	var deleted bool
	path := fmt.Sprintf("/session/%s", url.PathEscape(id))
	if err := c.doJSON(http.MethodDelete, path, nil, &deleted); err != nil {
		return false, err
	}

	return deleted, nil
}

func (c *OpencodeAgent[T]) sendMessage(id string, message core.TaskAgentMessage) (core.PromptResult[T], error) {
	slog.Info("req /session/{sessionID}/message", "action", "prompt", "session_id", id)

	part, err := newTextPromptPart(message.Text)
	if err != nil {
		return core.PromptResult[T]{}, err
	}

	body := SessionPromptJSONBody{
		Parts: []SessionPromptJSONBody_Parts_Item{part},
	}

	format, err := newOutputFormat[T]()
	if err != nil {
		return core.PromptResult[T]{}, err
	}
	body.Format = format

	path := fmt.Sprintf("/session/%s/message", url.PathEscape(id))
	var result core.PromptResult[T]
	if err := c.doJSON(http.MethodPost, path, body, &result); err != nil {
		return core.PromptResult[T]{}, err
	}

	return result, nil
}

func (c *OpencodeAgent[T]) prompt(message core.TaskAgentMessage) (core.PromptResult[T], error) {
	session, err := c.createSession()
	if err != nil {
		return core.PromptResult[T]{}, err
	}

	return c.sendMessage(session.Id, message)
}

func (c *OpencodeAgent[T]) doJSON(method, path string, requestBody any, responseBody any) error {
	endpoint, err := c.resolveURL(path)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if requestBody != nil {
		encodedBody, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal %s %s request: %w", method, endpoint, err)
		}
		bodyReader = bytes.NewReader(encodedBody)
	}

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("create %s %s request: %w", method, endpoint, err)
	}

	req.Header.Set("Accept", "application/json")
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send %s %s request: %w", method, endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return unexpectedStatusError(method, endpoint, resp)
	}

	if responseBody == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
		return fmt.Errorf("decode %s %s response: %w", method, endpoint, err)
	}

	return nil
}

func (c *OpencodeAgent[T]) resolveURL(path string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base url %q: %w", c.baseURL, err)
	}

	reference, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("parse path %q: %w", path, err)
	}

	return baseURL.ResolveReference(reference).String(), nil
}

func unexpectedStatusError(method, endpoint string, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s %s returned %s and response body could not be read: %w", method, endpoint, resp.Status, err)
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("%s %s returned %s", method, endpoint, resp.Status)
	}

	return fmt.Errorf("%s %s returned %s: %s", method, endpoint, resp.Status, message)
}
