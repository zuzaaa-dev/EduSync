package group

// Group представляет учебную группу
// swagger:model
type Group struct {
	// ID группы
	// example: 1
	ID int `json:"id"`

	// Название группы
	// example: ИС-42
	Name string `json:"name"`

	// ID учебного заведения
	// example: 5
	InstitutionID int `json:"institution_id"`
}
