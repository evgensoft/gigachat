package gigachat

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	baseURLToken = "https://ngw.devices.sberbank.ru:9443/api/v2"
	baseURL      = "https://gigachat.devices.sberbank.ru/api/v1"
)

// Client представляет клиент для работы с API GigaChat
type Client struct {
	clientID     string
	clientSecret string
	basicAuth    string
	scope        string
	httpClient   *http.Client
	token        *TokenResponse
	tokenExpiry  time.Time
}

// NewClient создает новый клиент GigaChat
func NewClient(clientID, clientSecret string) *Client {
	// Создаем базовую авторизацию
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		basicAuth:    fmt.Sprintf("Basic %s", auth),
		scope:        ScopePersonal, // По умолчанию используем персональный scope
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetScope устанавливает scope для клиента
func (c *Client) SetScope(scope string) {
	c.scope = scope
}

// GetToken получает токен доступа
func (c *Client) GetToken() (*TokenResponse, error) {
	dataScope := fmt.Sprintf("scope=%s", c.scope)

	// Создаем запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth", baseURLToken), bytes.NewBufferString(dataScope))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", c.basicAuth)
	httpReq.Header.Set("RqUID", generateUUID())

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("ошибка разбора ответа об ошибке: %w", err)
		}
		return nil, &APIError{
			Code:    errResp.Code,
			Message: errResp.Message,
		}
	}

	// Разбираем успешный ответ
	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	// Сохраняем токен и время его истечения
	c.token = &tokenResp
	c.tokenExpiry = time.Unix(tokenResp.ExpiresAt/1000-60, 0)

	return &tokenResp, nil
}

// ensureValidToken проверяет и обновляет токен при необходимости
func (c *Client) ensureValidToken() error {
	if c.token == nil || time.Now().After(c.tokenExpiry) {
		_, err := c.GetToken()
		if err != nil {
			return fmt.Errorf("ошибка обновления токена: %w", err)
		}
	}
	return nil
}

// Chat отправляет запрос к чату
func (c *Client) Chat(req *ChatRequest) (*ChatResponse, error) {
	// Проверяем валидность токена
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	// Создаем запрос
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))
	httpReq.Header.Set("X-Client-ID", c.clientID)
	httpReq.Header.Set("X-Request-ID", generateUUID())
	httpReq.Header.Set("X-Session-ID", generateUUID())

	// Отправляем запрос
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("ошибка разбора ответа об ошибке: %w", err)
		}
		return nil, &APIError{
			Code:    errResp.Code,
			Message: errResp.Message,
		}
	}

	// Разбираем успешный ответ
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
	}

	return &chatResp, nil
}

// generateUUID генерирует UUID для RqUID
func generateUUID() string {
	// Генерация реального UUID
	return uuid.New().String()
}
