package institution

// Institution представляет учебное заведение.
type Institution struct {
	ID   int
	Name string
}

// EmailMask представляет запись почтовой маски для учебного заведения.
type EmailMask struct {
	InstitutionID int    `json:"institution_id"`
	EmailMask     string `json:"email_mask"`
}
