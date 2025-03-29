package schedule

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"time"

	"EduSync/internal/integration/parser/rksi"

	"github.com/PuerkitoBio/goquery"
)

const (
	timeRangeSplit   = "—"
	infoSplit        = ","
	emptyPlaceholder = "-"
	dateLayout       = "02-01-2006"
)

// ScheduleEntry представляет одну запись расписания.
type ScheduleEntry struct {
	Date       time.Time
	StartTime  string
	EndTime    string
	Discipline string
	Teacher    string
	Classroom  string
}

var monthMapping = map[string]time.Month{
	"января":   time.January,
	"февраля":  time.February,
	"марта":    time.March,
	"апреля":   time.April,
	"мая":      time.May,
	"июня":     time.June,
	"июля":     time.July,
	"августа":  time.August,
	"сентября": time.September,
	"октября":  time.October,
	"ноября":   time.November,
	"декабря":  time.December,
}

type Parser interface {
	FetchSchedule(ctx context.Context, group string) ([]ScheduleEntry, error)
}

type ScheduleParser struct {
	URL           string
	log           *logrus.Logger
	InstitutionID int
}

func NewScheduleParser(URL string, log *logrus.Logger) *ScheduleParser {
	return &ScheduleParser{URL: URL, log: log, InstitutionID: 1}
}

// parseRussianDate преобразует русскую дату в `time.Time`.
func parseRussianDate(russianDate string) (time.Time, error) {
	parts := strings.SplitN(russianDate, ",", 2)
	dateParts := strings.Fields(strings.TrimSpace(parts[0]))
	if len(dateParts) != 2 {
		return time.Time{}, fmt.Errorf("неверный формат даты: %s", russianDate)
	}

	day := dateParts[0]
	monthStr := strings.ToLower(dateParts[1])

	month, ok := monthMapping[monthStr]
	if !ok {
		return time.Time{}, fmt.Errorf("неизвестный месяц: %s", monthStr)
	}

	now := time.Now()
	year := now.Year()

	// Учитываем переход на новый год
	if now.Month() == time.December && month == time.January {
		year++
	}

	return time.Parse(dateLayout, fmt.Sprintf("%02s-%02d-%d", day, month, year))
}

// ParseSchedule извлекает расписание из HTML.
func ParseSchedule(html []byte) ([]ScheduleEntry, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки HTML: %w", err)
	}

	var entries []ScheduleEntry

	doc.Find(".schedule_item").Each(func(i int, s *goquery.Selection) {
		dateStr := s.Find(".schedule_title").Text()
		date, err := parseRussianDate(dateStr)
		if err != nil {
			log.Printf("Ошибка парсинга даты: %v", err)
			return
		}

		s.Find("p").Each(func(j int, p *goquery.Selection) {
			var parts []string
			p.Contents().Each(func(k int, node *goquery.Selection) {
				if !node.Is("br") {
					text := strings.TrimSpace(node.Text())
					if text != "" {
						parts = append(parts, text)
					}
				}
			})

			if len(parts) < 2 {
				log.Printf("Недостаточно данных: %v", parts)
				return
			}

			timeParts := strings.Split(parts[0], timeRangeSplit)
			if len(timeParts) != 2 {
				log.Printf("Некорректный формат времени: %s", parts[0])
				return
			}

			entry := ScheduleEntry{
				Date:       date,
				StartTime:  strings.TrimSpace(timeParts[0]),
				EndTime:    strings.TrimSpace(timeParts[1]),
				Discipline: parts[1],
				Teacher:    emptyPlaceholder,
				Classroom:  emptyPlaceholder,
			}

			if len(parts) > 2 {
				infoParts := strings.SplitN(parts[2], infoSplit, 2)
				for i, part := range infoParts {
					value := strings.TrimSpace(part)
					switch i {
					case 0:
						if value != "" && !strings.HasPrefix(value, "_") {
							entry.Teacher = value
						}
					case 1:
						if value != "" {
							entry.Classroom = value
						}
					}
				}
			}

			entries = append(entries, entry)
		})
	})

	return entries, nil
}

// FetchSchedule получает и парсит расписание для группы.
func (s *ScheduleParser) FetchSchedule(ctx context.Context, group string) ([]ScheduleEntry, error) {
	html, err := rksi.FetchHTML(ctx, s.URL, map[string]string{"group": group, "stt": "Показать!"})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения расписания: %w", err)
	}
	return ParseSchedule(html)
}
