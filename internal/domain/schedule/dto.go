package schedule

import "time"

// Item — элемент расписания
// swagger:model
type Item struct {
	// ID записи расписания
	// example: 184
	ID int `json:"id"`

	// ID группы
	// example: 1
	GroupID int `json:"group_id"`

	// Название группы
	// example: 10-Б
	GroupName string `json:"group_name"`

	// Название предмета
	// example: Математика
	Subject string `json:"subject"`

	// Дата занятия
	// example: 2025-03-31T00:00:00Z
	Date time.Time `json:"date"`

	// Номер пары
	// example: 1
	PairNumber int `json:"pair_number"`

	// Аудитория
	// example: ауд. 204
	Classroom string `json:"classroom"`

	// ID преподавателя (если он зарегистрирован)
	// example: 42
	TeacherID *int `json:"teacher_id,omitempty"`

	// Инициалы преподавателя
	// example: И.И.Иванов
	TeacherInitials string `json:"teacher_initials,omitempty"`

	// Время начала пары
	// example: 0001-01-01T08:00:00Z
	StartTime time.Time `json:"start_time"`

	// Время конца пары
	// example: 0001-01-01T09:30:00Z
	EndTime time.Time `json:"end_time"`
}
