package models

type SaveUserPayload struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

type EditUserPayload struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

type DeleteUserPayload struct {
	ID int64 `json:"id"`
}

type Country struct {
	CountryID   string  `json:"country_id"`
	Probability float64 `json:"probability"`
}

type EnrichedUser struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname,omitempty"`
	Patronymic string    `json:"patronymic"`
	Age        int       `json:"age"`
	Sex        string    `json:"sex"`
	Country    []Country `json:"country"`
}

type UserFilter struct {
	// Поиск по имени (частичное совпадение)
	Name string `json:"name,omitempty"`

	// Поиск по фамилии (частичное совпадение)
	Surname string `json:"surname,omitempty"`

	// Поиск по отчеству (частичное совпадение)
	Patronymic string `json:"patronymic,omitempty"`

	// Возраст от (включительно)
	AgeFrom int `json:"ageFrom,omitempty"`

	// Возраст до (включительно)
	AgeTo int `json:"ageTo,omitempty"`

	// Пол (точное совпадение)
	Sex string `json:"sex,omitempty"`

	// Страна (частичное совпадение)
	Country string `json:"country,omitempty"`

	// Пагинация - количество записей на странице
	Limit int `json:"limit,omitempty"`

	// Пагинация - смещение (пропуск записей)
	Offset int `json:"offset,omitempty"`
}
