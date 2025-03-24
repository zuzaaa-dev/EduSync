package group

// Group представляет доменную модель группы.
type Group struct {
	ID            int    // Может быть нулевым, если еще не сохранена в БД
	Name          string // Название группы, полученное с сайта
	InstitutionID int    // ID учебного заведения
}
