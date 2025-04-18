package schedule

// TeacherInitialDTO — то, что отдадим по API.
type TeacherInitialDTO struct {
	ID       int    `json:"id"`
	Initials string `json:"initials"`
}
