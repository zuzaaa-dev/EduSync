package group

import (
	"EduSync/internal/integration/parser/rksi"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
)

// GroupParser предоставляет методы для парсинга групп.
type GroupParser struct {
	URL string
}

// NewGroupParser создает новый парсер групп.
func NewGroupParser(url string) *GroupParser {
	return &GroupParser{URL: url}
}

// FetchGroups получает список групп с сайта.
func (p *GroupParser) FetchGroups() ([]string, error) {
	html, err := rksi.FetchPage(p.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения страницы: %v", err)
	}

	return parseGroups(html)
}

// parseGroups обрабатывает HTML и извлекает группы.
func parseGroups(html []byte) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки HTML: %v", err)
	}

	var groups []string
	doc.Find("select#group option").Each(func(i int, s *goquery.Selection) {
		groupName, _ := s.Attr("value")
		if groupName != "" && groupName != "_" { // Игнорируем пустые значения
			groups = append(groups, groupName)
		}
	})

	return groups, nil
}
