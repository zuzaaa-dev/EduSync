package rksi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Parser описывает интерфейс для получения групп.
type Parser interface {
	FetchGroups() ([]string, error)
}

// FetchPage выполняет запрос к сайту и возвращает HTML-страницу.
func FetchPage(targetURL string, params map[string]string) ([]byte, error) {
	client := &http.Client{}

	// Формируем тело POST-запроса, если есть параметры
	var requestBody *bytes.Reader
	if len(params) > 0 {
		formData := url.Values{}
		for key, value := range params {
			formData.Set(key, value)
		}
		requestBody = bytes.NewReader([]byte(formData.Encode()))
	} else {
		requestBody = bytes.NewReader(nil)
	}

	// Создаем HTTP-запрос
	req, err := http.NewRequest("POST", targetURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неожиданный статус ответа: %d", resp.StatusCode)
	}

	// Читаем тело ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}

	return body, nil
}
