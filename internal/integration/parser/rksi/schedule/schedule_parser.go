package schedule

import (
	"EduSync/internal/integration/parser/rksi"
	"bytes"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ScheduleEntry представляет одну запись расписания.
type ScheduleEntry struct {
	Date       string
	StartTime  string
	EndTime    string
	Discipline string
	Teacher    string
	Classroom  string
}

// ScheduleParser предоставляет методы для парсинга расписания.
type ScheduleParser struct {
	URL string
}

// NewScheduleParser создает новый парсер расписания.
func NewScheduleParser(url string) *ScheduleParser {
	return &ScheduleParser{URL: url}
}

// FetchSchedule получает расписание для конкретной группы.
func (p *ScheduleParser) FetchSchedule(group string) ([]ScheduleEntry, error) {
	html, err := rksi.FetchPage(p.URL, map[string]string{"group": group, "stt": "Показать!"})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения страницы: %v", err)
	}

	return parseSchedule(html)
}

// parseSchedule обрабатывает HTML и извлекает расписание.
func parseSchedule(html []byte) ([]ScheduleEntry, error) {
	var entries []ScheduleEntry

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки HTML: %v", err)
	}

	// Проходим по элементам расписания
	doc.Find(".schedule_item").Each(func(i int, scheduleItem *goquery.Selection) {
		date := scheduleItem.Find(".schedule_title").Text()

		// Извлекаем занятия
		scheduleItem.Find("p").Each(func(j int, p *goquery.Selection) {
			var parts []string

			p.Contents().Each(func(k int, node *goquery.Selection) {
				if !node.Is("br") {
					text := strings.TrimSpace(node.Text())
					if text != "" {
						parts = append(parts, text)
					}
				}
			})

			// Если данных недостаточно — пропускаем
			if len(parts) < 2 {
				return
			}

			// Разбираем время
			timeRange := parts[0]
			timeParts := strings.Split(timeRange, "—")
			if len(timeParts) != 2 {
				return
			}
			startTime := strings.TrimSpace(timeParts[0])
			endTime := strings.TrimSpace(timeParts[1])

			// Название дисциплины
			discipline := parts[1]

			// Преподаватель и аудитория
			teacher := "-"
			classroom := "-"
			if len(parts) > 2 {
				infoParts := strings.Split(parts[2], ",")
				if len(infoParts) == 2 {
					teacher = strings.TrimSpace(infoParts[0])
					classroom = strings.TrimSpace(infoParts[1])
				} else if len(infoParts) == 1 {
					classroom = strings.TrimSpace(infoParts[0])
				}
			}

			entry := ScheduleEntry{
				Date:       date,
				StartTime:  startTime,
				EndTime:    endTime,
				Discipline: discipline,
				Teacher:    teacher,
				Classroom:  classroom,
			}

			entries = append(entries, entry)
		})
	})

	return entries, nil
}
