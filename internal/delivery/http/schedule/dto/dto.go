package dto

import "time"

// TeacherInitialDTO — то, что отдадим по API.
type TeacherInitialDTO struct {
	ID       int    `json:"id"`
	Initials string `json:"initials"`
}

// UpdateScheduleReq — все поля опциональные, кроме ID в пути.
type UpdateScheduleReq struct {
	GroupId           *int       `json:"group_id,omitempty"`
	SubjectID         *int       `json:"subject_id,omitempty"`
	Date              *time.Time `json:"date,omitempty"`
	PairNumber        *int       `json:"pair_number,omitempty"`
	Classroom         *string    `json:"classroom,omitempty"`
	TeacherInitialsID *int       `json:"teacher_initials_id,omitempty"`
	StartTime         *time.Time `json:"start_time,omitempty"`
	EndTime           *time.Time `json:"end_time,omitempty"`
}
