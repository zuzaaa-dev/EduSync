package schedule

import "time"

// Schedule представляет доменную модель записи расписания.
type Schedule struct {
	ID              int       `json:"id"`
	GroupID         int       `json:"group_id"`
	SubjectID       int       `json:"subject_id"`
	Date            time.Time `json:"date"`
	PairNumber      int       `json:"pair_number"`
	Classroom       string    `json:"classroom"`
	TeacherID       *int      `json:"teacher_id"`       // Может быть NULL
	TeacherInitials string    `json:"teacher_initials"` // Если teacher_id не найден
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
}
