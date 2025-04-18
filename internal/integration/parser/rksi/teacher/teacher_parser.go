package teacher

import (
	"EduSync/internal/integration/parser/rksi"
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
)

type Parser interface {
	FetchTeacher(ctx context.Context) ([]string, int, error)
}

type TeacherParser struct {
	URL string
	log *logrus.Logger
}

func NewTeacherParser(URL string, log *logrus.Logger) *TeacherParser {
	return &TeacherParser{URL: URL, log: log}
}

// ParseTeacher извлекает список преподавателей из HTML-кода.
func ParseTeacher(html []byte) ([]string, int, error) {
	const (
		groupSelector = "select#teacher option"
		groupAttr     = "value"
	)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка парсинга HTML: %w", err)
	}

	var teachers []string
	doc.Find(groupSelector).Each(func(i int, s *goquery.Selection) {
		if groupName, exists := s.Attr(groupAttr); exists {
			teachers = append(teachers, groupName)
		}
	})

	if len(teachers) == 0 {
		return nil, 0, fmt.Errorf("преподаватели не найдены")
	}

	return teachers, 1, nil
}

// FetchTeacher получает и парсит расписание для группы.
func (s *TeacherParser) FetchTeacher(ctx context.Context) ([]string, int, error) {
	html, err := rksi.FetchHTML(ctx, s.URL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения расписания: %w", err)
	}
	return ParseTeacher(html)
}
