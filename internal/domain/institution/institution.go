package institution

// Institution представляет учебное заведение
// swagger:model
type Institution struct {
	// ID учреждения
	// example: 1
	ID int `json:"id"`

	// Название учреждения
	// example: Московский Политех
	Name string `json:"name"`
}

// EmailMask представляет почтовую маску
// swagger:model
type EmailMask struct {
	// ID учебного заведения
	// example: 5
	InstitutionID int `json:"institution_id"`

	// Маска email
	// example: "@mospolytech.ru"
	EmailMask string `json:"email_mask"`
}
