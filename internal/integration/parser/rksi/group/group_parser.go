package group

import (
	"EduSync/internal/integration/parser/rksi"
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
)

type Parser interface {
	FetchGroups(ctx context.Context) ([]string, int, error)
}

type GroupParser struct {
	URL string
	log *logrus.Logger
}

func NewGroupParser(URL string, log *logrus.Logger) Parser {
	return &GroupParser{URL: URL, log: log}
}

// ParseGroups извлекает список групп из HTML-кода.
func ParseGroups(html []byte) ([]string, int, error) {
	const (
		groupSelector = "select#group option"
		groupAttr     = "value"
	)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка парсинга HTML: %w", err)
	}

	var groups []string
	doc.Find(groupSelector).Each(func(i int, s *goquery.Selection) {
		if groupName, exists := s.Attr(groupAttr); exists {
			groups = append(groups, groupName)
		}
	})

	if len(groups) == 0 {
		return nil, 0, fmt.Errorf("группы не найдены")
	}

	return groups, 1, nil
}

// FetchGroups получает и парсит расписание для группы.
func (s *GroupParser) FetchGroups(ctx context.Context) ([]string, int, error) {
	html, err := rksi.FetchHTML(ctx, s.URL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения расписания: %w", err)
	}
	return ParseGroups(html)
}
