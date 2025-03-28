package subject

// Subject представляет предмет, связанный с учебным заведением.
type Subject struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	InstitutionID int    `json:"institution_id"`
}
