package group

import (
	"EduSync/internal/integration/parser/rksi"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// GroupParser предоставляет методы для парсинга групп.
type GroupParser struct {
	URL string
	log *logrus.Logger
}

// NewGroupParser создает новый парсер групп.
func NewGroupParser(url string, logger *logrus.Logger) *GroupParser {
	return &GroupParser{
		URL: url,
		log: logger,
	}
}

// FetchGroups получает список групп с сайта.
func (p *GroupParser) FetchGroups() ([]string, int, error) {
	p.log.Infof("Начато отслеживание групп: %s", p.URL)
	html, err := rksi.FetchPage(p.URL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения страницы: %v", err)
	}

	return parseGroups(html, p.log)
}

// parseGroups обрабатывает HTML и извлекает группы.
func parseGroups(html []byte, log *logrus.Logger) ([]string, int, error) {
	log.Infof("Парсинг групп: %s", html)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка загрузки HTML: %v", err)
	}

	var groups []string
	doc.Find("select#group option").Each(func(i int, s *goquery.Selection) {
		groupName, _ := s.Attr("value")
		if groupName != "" && groupName != "_" { // Игнорируем пустые значения
			groups = append(groups, groupName)
		}
	})

	return groups, 1, nil
}
