package schedule

import "time"

// Schedule представляет доменную модель записи расписания.
type Schedule struct {
	ID                int       `json:"id"`
	GroupID           int       `json:"group_id"`
	SubjectID         int       `json:"subject_id"`
	Date              time.Time `json:"date"`
	PairNumber        int       `json:"pair_number"`
	Classroom         string    `json:"classroom"`
	TeacherInitialsID *int      `json:"teacher_initials_id,omitempty"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
}

type TeacherInitials struct {
	ID            int    `json:"id"`
	Initials      string `json:"initials"`
	TeacherID     *int   `json:"teacher_id,omitempty"`
	InstitutionID int    `json:"institution_id"`
}
