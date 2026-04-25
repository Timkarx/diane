package opencode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Timkarx/diane/core"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
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

func newOutputFormat[S core.TaskAgentMessage, K any, T core.TaskSpec[S, K]]() (*OutputFormat, error) {
	var action T
	schema := action.Schema()
	if len(schema) == 0 {
		return nil, nil
	}

	var outputFormat OutputFormat
	err := outputFormat.FromOutputFormatJsonSchema(OutputFormatJsonSchema{
		Schema: JSONSchema(schema),
		Type:   "json_schema",
	})
	if err != nil {
		return nil, fmt.Errorf("encode response format: %w", err)
	}

	return &outputFormat, nil
}

func (c *OpencodeAgent[S, K, T]) createSession() (Session, error) {
	slog.Info("req /session", "action", "create")

	title := fmt.Sprintf("analyze_%d", atomic.AddUint64(&c.requestCounter, 1)-1)
	body := SessionCreateJSONBody{Title: &title}

	var session Session
	if err := c.doJSON(http.MethodPost, "/session", body, &session); err != nil {
		return Session{}, err
	}

	return session, nil
}

func (c *OpencodeAgent[S, K, T]) deleteSession(id string) (bool, error) {
	slog.Info("req /session/{sessionID}", "action", "delete", "session_id", id)

	var deleted bool
	path := fmt.Sprintf("/session/%s", url.PathEscape(id))
	if err := c.doJSON(http.MethodDelete, path, nil, &deleted); err != nil {
		return false, err
	}

	return deleted, nil
}

func (c *OpencodeAgent[S, K, T]) sendMessage(id string, message string) (OpencodeResult[K], error) {
	slog.Info("req /session/{sessionID}/message", "action", "prompt", "session_id", id)

	part, err := newTextPromptPart(message)
	if err != nil {
		return OpencodeResult[K]{}, err
	}

	body := SessionPromptJSONBody{
		Parts: []SessionPromptJSONBody_Parts_Item{part},
	}

	agent, model := c.promptOptions()
	body.Agent = agent
	if model != nil {
		body.Model = &struct {
			ModelID    string `json:"modelID"`
			ProviderID string `json:"providerID"`
		}{
			ModelID:    model.modelID,
			ProviderID: model.providerID,
		}
	}

	format, err := newOutputFormat[S, K, T]()
	if err != nil {
		return OpencodeResult[K]{}, err
	}
	body.Format = format

	path := fmt.Sprintf("/session/%s/message", url.PathEscape(id))
	var result OpencodeResult[K]
	if err := c.doJSON(http.MethodPost, path, body, &result); err != nil {
		return OpencodeResult[K]{}, err
	}

	return result, nil
}

func (c *OpencodeAgent[S, K, T]) prompt(message string) (OpencodeResult[K], error) {
	if c.sessionMode == core.TaskAgentSessionModeReusePerClient {
		return c.promptWithReusableSession(message)
	}

	session, err := c.createSession()
	if err != nil {
		return OpencodeResult[K]{}, err
	}

	return c.sendMessage(session.Id, message)
}

func (c *OpencodeAgent[S, K, T]) promptWithReusableSession(message string) (OpencodeResult[K], error) {
	sessionID, err := c.getOrCreateReusableSessionID()
	if err != nil {
		return OpencodeResult[K]{}, err
	}

	result, err := c.sendMessage(sessionID, message)
	if err == nil || !isSessionNotFoundError(err) {
		return result, err
	}

	c.clearReusableSessionID(sessionID)

	sessionID, err = c.getOrCreateReusableSessionID()
	if err != nil {
		return OpencodeResult[K]{}, err
	}

	return c.sendMessage(sessionID, message)
}

func (c *OpencodeAgent[S, K, T]) getOrCreateReusableSessionID() (string, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()

	if c.sessionID != "" {
		return c.sessionID, nil
	}

	session, err := c.createSession()
	if err != nil {
		return "", err
	}

	c.sessionID = session.Id
	return c.sessionID, nil
}

func (c *OpencodeAgent[S, K, T]) clearReusableSessionID(sessionID string) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()

	if c.sessionID == sessionID {
		c.sessionID = ""
	}
}

func (c *OpencodeAgent[S, K, T]) doJSON(method, path string, requestBody any, responseBody any) error {
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

func (c *OpencodeAgent[S, K, T]) resolveURL(path string) (string, error) {
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
		return &statusError{
			method:     method,
			endpoint:   endpoint,
			status:     resp.Status,
			statusCode: resp.StatusCode,
			readErr:    err,
		}
	}

	message := strings.TrimSpace(string(body))
	return &statusError{
		method:     method,
		endpoint:   endpoint,
		status:     resp.Status,
		statusCode: resp.StatusCode,
		message:    message,
	}
}

type statusError struct {
	method     string
	endpoint   string
	status     string
	statusCode int
	message    string
	readErr    error
}

func (e *statusError) Error() string {
	if e.readErr != nil {
		return fmt.Sprintf("%s %s returned %s and response body could not be read: %v", e.method, e.endpoint, e.status, e.readErr)
	}

	if e.message == "" {
		return fmt.Sprintf("%s %s returned %s", e.method, e.endpoint, e.status)
	}

	return fmt.Sprintf("%s %s returned %s: %s", e.method, e.endpoint, e.status, e.message)
}

func isSessionNotFoundError(err error) bool {
	var statusErr *statusError
	return errors.As(err, &statusErr) && statusErr.statusCode == http.StatusNotFound
}
