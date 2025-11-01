package models

type Product struct {
	ID uint `json:"id" gorm:"primaryKey;column:id"`
}
