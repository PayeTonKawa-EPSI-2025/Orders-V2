package models

type Customer struct {
	ID uint `json:"id" gorm:"primaryKey;column:id"`
}
