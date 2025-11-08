package operation

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/PayeTonKawa-EPSI-2025/Common/events"
	"github.com/PayeTonKawa-EPSI-2025/Common/models"
	"github.com/PayeTonKawa-EPSI-2025/Orders/internal/dto"
	localModels "github.com/PayeTonKawa-EPSI-2025/Orders/internal/models"
	"github.com/PayeTonKawa-EPSI-2025/Orders/internal/rabbitmq"
	"github.com/danielgtaylor/huma/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

func RegisterOrdersRoutes(api huma.API, dbConn *gorm.DB, ch *amqp.Channel) {

	huma.Register(api, huma.Operation{
		OperationID: "get-orders",
		Summary:     "Get all orders",
		Method:      http.MethodGet,
		Path:        "/orders",
		Tags:        []string{"orders"},
	}, func(ctx context.Context, input *struct{}) (*dto.OrdersOutput, error) {
		resp := &dto.OrdersOutput{}

		var orders []models.Order
		results := dbConn.Find(&orders)

		if results.Error == nil {
			resp.Body.Orders = orders
		}

		return resp, results.Error
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-order",
		Summary:     "Get a order",
		Method:      http.MethodGet,
		Path:        "/orders/{id}",
		Tags:        []string{"orders"},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
	}) (*dto.OrderOutput, error) {
		resp := &dto.OrderOutput{}

		var order models.Order
		results := dbConn.First(&order, input.Id)

		if results.Error == nil {
			resp.Body = order
			return resp, nil
		}

		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Order not found")
		}

		return nil, results.Error
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-order",
		Summary:       "Create an order",
		Method:        http.MethodPost,
		DefaultStatus: http.StatusCreated,
		Path:          "/orders",
		Tags:          []string{"orders"},
	}, func(ctx context.Context, input *dto.OrderCreateInput) (*dto.OrderOutput, error) {
		resp := &dto.OrderOutput{}

		order := models.Order{
			CustomerID: input.Body.CustomerID,
			Products:   input.Body.Products,
		}

		// Create order in the database
		results := dbConn.Create(&order)
		if results.Error != nil {
			return resp, results.Error
		}

		// Create CustomerOrder relationship
		customerOrder := localModels.CustomerOrder{
			CustomerID: input.Body.CustomerID,
			OrderID:    order.ID,
		}

		if err := dbConn.Create(&customerOrder).Error; err != nil {
			// Log this but don't fail the order creation itself
			fmt.Printf("Failed to create CustomerOrder record: %v\n", err)
		}

		// Prepare response
		resp.Body = order

		// Publish order created event
		if err := rabbitmq.PublishOrderEvent(ch, events.OrderCreated, order); err != nil {
			// Log error but do not fail the request
			fmt.Printf("Failed to publish order event: %v\n", err)
		}

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "put-order",
		Summary:     "Replace a order",
		Method:      http.MethodPut,
		Path:        "/orders/{id}",
		Tags:        []string{"orders"},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
		dto.OrderCreateInput
	}) (*dto.OrderOutput, error) {
		resp := &dto.OrderOutput{}

		var order models.Order
		results := dbConn.First(&order, input.Id)

		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Order not found")
		}
		if results.Error != nil {
			return nil, results.Error
		}

		updates := models.Order{
			CustomerID: input.Body.CustomerID,
			Products:   input.Body.Products,
		}

		results = dbConn.Model(&order).Updates(updates)
		if results.Error != nil {
			return nil, results.Error
		}

		// Get updated order from DB to ensure all fields are correct
		dbConn.First(&order, order.ID)
		resp.Body = order

		// Publish order updated event
		err := rabbitmq.PublishOrderEvent(ch, events.OrderUpdated, order)
		if err != nil {
			// Log the error but don't fail the request
			// The order was already updated in the database
		}

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID:   "delete-order",
		Summary:       "Delete a order",
		Method:        http.MethodDelete,
		DefaultStatus: http.StatusNoContent,
		Path:          "/orders/{id}",
		Tags:          []string{"orders"},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
	}) (*struct{}, error) {
		resp := &struct{}{}

		// First get the order to have the complete data for the event
		var order models.Order
		result := dbConn.First(&order, input.Id)

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Order not found")
		}

		if result.Error != nil {
			return nil, result.Error
		}

		results := dbConn.Delete(&order)

		if results.Error == nil {
			// Publish order deleted event
			err := rabbitmq.PublishOrderEvent(ch, events.OrderDeleted, order)
			if err != nil {
				// Log the error but don't fail the request
				// The order was already deleted from the database
			}

			return resp, nil
		}

		return nil, results.Error
	})
}
