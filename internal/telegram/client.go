package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Client struct {
	token  string
	apiURL string
	http   *http.Client
}

type Update struct {
	UpdateID         int               `json:"update_id"`
	Message          *Message          `json:"message"`
	CallbackQuery    *CallbackQuery    `json:"callback_query"`
	PreCheckoutQuery *PreCheckoutQuery `json:"pre_checkout_query"`
}

type Message struct {
	MessageID         int                `json:"message_id"`
	From              *User              `json:"from"`
	Chat              Chat               `json:"chat"`
	Text              string             `json:"text"`
	SuccessfulPayment *SuccessfulPayment `json:"successful_payment"`
}

type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type PreCheckoutQuery struct {
	ID             string `json:"id"`
	From           User   `json:"from"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
	InvoicePayload string `json:"invoice_payload"`
}

type SuccessfulPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int    `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type LabeledPrice struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

func NewClient(token string) *Client {
	return &Client{
		token:  token,
		apiURL: "https://api.telegram.org/bot" + token,
		http: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *Client) GetUpdates(ctx context.Context, offset int, timeout int) ([]Update, error) {
	params := url.Values{}
	params.Set("timeout", strconv.Itoa(timeout))
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}

	var response struct {
		OK          bool     `json:"ok"`
		Description string   `json:"description"`
		Result      []Update `json:"result"`
	}
	if err := c.get(ctx, "getUpdates", params, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", response.Description)
	}
	return response.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	return c.postJSON(ctx, "sendMessage", payload, nil)
}

func (c *Client) SendMessageWithKeyboard(ctx context.Context, chatID int64, text string, keyboard InlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":      chatID,
		"text":         text,
		"reply_markup": keyboard,
	}
	return c.postJSON(ctx, "sendMessage", payload, nil)
}

func (c *Client) EditMessageText(ctx context.Context, chatID int64, messageID int, text string, keyboard InlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":      chatID,
		"message_id":   messageID,
		"text":         text,
		"reply_markup": keyboard,
	}
	return c.postJSON(ctx, "editMessageText", payload, nil)
}

func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackID, text string) error {
	payload := map[string]any{
		"callback_query_id": callbackID,
		"text":              text,
	}
	return c.postJSON(ctx, "answerCallbackQuery", payload, nil)
}

func (c *Client) SendInvoice(ctx context.Context, chatID int64, title, description, payload string, stars int) error {
	body := map[string]any{
		"chat_id":        chatID,
		"title":          title,
		"description":    description,
		"payload":        payload,
		"provider_token": "",
		"currency":       "XTR",
		"prices": []LabeledPrice{
			{Label: title, Amount: stars},
		},
	}
	return c.postJSON(ctx, "sendInvoice", body, nil)
}

func (c *Client) AnswerPreCheckoutQuery(ctx context.Context, queryID string, ok bool, errorMessage string) error {
	payload := map[string]any{
		"pre_checkout_query_id": queryID,
		"ok":                    ok,
	}
	if !ok && errorMessage != "" {
		payload["error_message"] = errorMessage
	}
	return c.postJSON(ctx, "answerPreCheckoutQuery", payload, nil)
}

func (c *Client) SendVideo(ctx context.Context, chatID int64, path string, caption string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		return err
	}
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	part, err := writer.CreateFormFile("video", filepath.Base(path))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/sendVideo", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram sendVideo failed: %s", apiResp.Description)
	}
	return nil
}

func (c *Client) get(ctx context.Context, method string, params url.Values, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL+"/"+method+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram %s returned HTTP %d", method, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(into)
}

func (c *Client) postJSON(ctx context.Context, method string, payload any, into any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/"+method, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp struct {
		OK          bool            `json:"ok"`
		Description string          `json:"description"`
		Result      json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return err
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram %s failed: %s", method, apiResp.Description)
	}
	if into != nil && len(apiResp.Result) > 0 {
		return json.Unmarshal(apiResp.Result, into)
	}
	return nil
}
