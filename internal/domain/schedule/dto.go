package schedule

import "time"

// swagger:model
type Item struct {
	ID              int       `json:"id"`
	GroupID         int       `json:"group_id"`
	Subject         string    `json:"subject"`
	Date            time.Time `json:"date"`
	PairNumber      int       `json:"pair_number"`
	Classroom       string    `json:"classroom"`
	TeacherID       *int      `json:"teacher_id,omitempty"`
	TeacherInitials string    `json:"teacher_initials,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
}
