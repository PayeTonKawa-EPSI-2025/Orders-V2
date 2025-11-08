package models

import "github.com/PayeTonKawa-EPSI-2025/Common/models"

type OrderProduct struct {
	ID        uint         `json:"id" gorm:"primaryKey"`
	OrderID   uint         `json:"orderId"`
	Order     models.Order `gorm:"foreignKey:OrderID"`
	ProductID uint         `json:"productId"`
	Product   Product      `gorm:"foreignKey:ProductID"`
}
