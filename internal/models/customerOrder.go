package models

import "github.com/PayeTonKawa-EPSI-2025/Common/models"

type CustomerOrder struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	CustomerID uint         `json:"customerId"`
	Customer   Customer     `gorm:"foreignKey:CustomerID"`
	OrderID    uint         `json:"orderId"`
	Order      models.Order `gorm:"foreignKey:OrderID"`
}
