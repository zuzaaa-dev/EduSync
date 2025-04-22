package subject

// Subject модель учебного предмета
// swagger:model
type Subject struct {
	// ID предмета
	// example: 1
	ID int `json:"id"`

	// Название предмета
	// example: Математический анализ
	Name string `json:"name"`

	// ID учебного заведения
	// example: 5
	InstitutionID int `json:"institution_id"`
}
