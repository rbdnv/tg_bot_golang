package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"project/lib/e"
)

type Client struct {
	baseURL *url.URL
	client  http.Client
}

const (
	getUpdatesMethod  = "getUpdates"
	sendMessageMethod = "sendMessage"
	defaultTimeout    = 10 * time.Second
)

type apiResponse struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
}

func New(apiBase string, token string) *Client {
	return &Client{
		baseURL: newBaseURL(apiBase, token),
		client:  http.Client{Timeout: defaultTimeout},
	}
}

func newBaseURL(apiBase string, token string) *url.URL {
	apiBase = strings.TrimSpace(apiBase)
	if apiBase == "" {
		apiBase = "https://api.telegram.org"
	} else if !strings.Contains(apiBase, "://") {
		apiBase = "https://" + apiBase
	}

	u, err := url.Parse(apiBase)
	if err != nil {
		panic(fmt.Sprintf("invalid telegram api base url %q: %v", apiBase, err))
	}

	u.Path = path.Join(strings.TrimSuffix(u.Path, "/"), "bot"+token)
	return u
}

func (c *Client) Updates(ctx context.Context, offset int, limit int) (updates []Update, err error) {
	defer func() { err = e.WrapIfErr("can't do updates", err) }()

	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.doGetRequest(ctx, getUpdatesMethod, q)
	if err != nil {
		return nil, err
	}

	var res apiResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if !res.Ok {
		return nil, telegramAPIError(res)
	}

	var updatesResponse []Update
	if err := json.Unmarshal(res.Result, &updatesResponse); err != nil {
		return nil, err
	}

	return updatesResponse, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int, text string) error {
	form := url.Values{}
	form.Add("chat_id", strconv.Itoa(chatID))
	form.Add("text", text)

	data, err := c.doPostFormRequest(ctx, sendMessageMethod, form)
	if err != nil {
		return e.Wrap("can't send message", err)
	}

	var res apiResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return e.Wrap("can't send message", err)
	}

	if !res.Ok {
		return e.Wrap("can't send message", telegramAPIError(res))
	}

	return nil
}

func (c *Client) doGetRequest(ctx context.Context, method string, query url.Values) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, method, query, nil, "")
}

func (c *Client) doPostFormRequest(ctx context.Context, method string, form url.Values) ([]byte, error) {
	return c.doRequest(
		ctx,
		http.MethodPost,
		method,
		nil,
		strings.NewReader(form.Encode()),
		"application/x-www-form-urlencoded",
	)
}

func (c *Client) doRequest(ctx context.Context, httpMethod string, method string, query url.Values, body io.Reader, contentType string) (data []byte, err error) {
	defer func() { err = e.WrapIfErr("can't do request", err) }()

	u := *c.baseURL
	u.Path = path.Join(c.baseURL.Path, method)

	req, err := http.NewRequestWithContext(ctx, httpMethod, u.String(), body)
	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		req.URL.RawQuery = query.Encode()
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("telegram api returned status %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func telegramAPIError(res apiResponse) error {
	if res.Description == "" {
		return fmt.Errorf("telegram api returned ok=false")
	}

	if res.ErrorCode == 0 {
		return fmt.Errorf("telegram api returned ok=false: %s", res.Description)
	}

	return fmt.Errorf("telegram api returned ok=false (%d): %s", res.ErrorCode, res.Description)
}
