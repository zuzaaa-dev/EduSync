package rksi

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 10 * time.Second
)

// FetchHTML выполняет HTTP-запрос к сайту и возвращает HTML-страницу.
func FetchHTML(ctx context.Context, URL string, params map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var req *http.Request
	var err error

	if len(params) == 0 {
		req, err = http.NewRequestWithContext(ctx, "GET", URL, nil)
	} else {
		formData := url.Values{}
		for key, value := range params {
			formData.Set(key, value)
		}

		req, err = http.NewRequestWithContext(ctx, "POST", URL, strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неожиданный статус-код: %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	return buf.Bytes(), nil
}
