package dto

import "time"

// TeacherInitialDTO — модель данных для инициалов преподавателя
// swagger:model
type TeacherInitialDTO struct {
	// ID инициалов
	// example: 1
	ID int `json:"id"`

	// Инициалы преподавателя
	// example: "И.И. Иванов"
	Initials string `json:"initials"`
}

// UpdateScheduleReq — тело запроса для обновления расписания
// swagger:model
type UpdateScheduleReq struct {
	// ID группы
	// example: 1
	GroupId *int `json:"group_id,omitempty"`

	// ID предмета
	// example: 5
	SubjectID *int `json:"subject_id,omitempty"`

	// Дата занятия (RFC3339)
	// example: "2023-12-25T00:00:00Z"
	Date *time.Time `json:"date,omitempty"`

	// Номер пары
	// example: 3
	PairNumber *int `json:"pair_number,omitempty"`

	// Аудитория
	// example: "А-101"
	Classroom *string `json:"classroom,omitempty"`

	// ID инициалов преподавателя
	// example: 2
	TeacherInitialsID *int `json:"teacher_initials_id,omitempty"`

	// Время начала (RFC3339)
	// example: "2023-12-25T09:00:00Z"
	StartTime *time.Time `json:"start_time,omitempty"`

	// Время окончания (RFC3339)
	// example: "2023-12-25T10:30:00Z"
	EndTime *time.Time `json:"end_time,omitempty"`
}
