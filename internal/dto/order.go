package dto

import "github.com/PayeTonKawa-EPSI-2025/Common/models"

type OrdersOutput struct {
	Body struct {
		Orders []models.Order `json:"orders"`
	}
}

type OrderOutput struct {
	Body models.Order
}

type OrderCreateInput struct {
	Body struct {
		CustomerID uint             `json:"customerId"`
		Products   []models.Product `json:"products,omitempty"`
	}
}

type CustomerOrdersInput struct {
	CustomerID uint `json:"customerId" path:"customerId"`
}
