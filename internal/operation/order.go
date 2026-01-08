package operation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PayeTonKawa-EPSI-2025/Common-V2/auth"
	"github.com/PayeTonKawa-EPSI-2025/Common-V2/events"
	"github.com/PayeTonKawa-EPSI-2025/Common-V2/models"
	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/dto"
	localModels "github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/models"
	"github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/rabbitmq"
	"github.com/danielgtaylor/huma/v2"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

// ----------------------
// Helper Functions
// ----------------------

func isAdmin(claims *auth.Claims) bool {
	for _, role := range claims.Roles {
		if role == "admin" {
			return true
		}
	}
	return false
}

func checkOrderOwnership(db *gorm.DB, orderID uint, customerID uint) bool {
	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		return false
	}
	return order.CustomerID == customerID
}

// ----------------------
// Extracted CRUD Functions
// ----------------------

// Get all orders
func GetOrders(ctx context.Context, db *gorm.DB) (*dto.OrdersOutput, error) {

	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, huma.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	resp := &dto.OrdersOutput{}

	var orders []models.Order
	if isAdmin(claims) {
		// Admin can see all orders
		results := db.Find(&orders)
		if results.Error != nil {
			return nil, results.Error
		}
	} else {
		// User can only see their own orders
		results := db.Where("customer_id = ?", claims.PreferredUsername).Find(&orders)
		if results.Error != nil {
			return nil, results.Error
		}
	}

	resp.Body.Orders = orders

	return resp, nil
}

// Get a single order by ID
func GetOrder(ctx context.Context, db *gorm.DB, id uint) (*dto.OrderOutput, error) {
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, huma.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	resp := &dto.OrderOutput{}

	var order models.Order
	results := db.First(&order, id)

	if results.Error != nil {
		if errors.Is(results.Error, gorm.ErrRecordNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Order not found")
		}
		return nil, results.Error
	}

	if !isAdmin(claims) && fmt.Sprintf("%d", order.CustomerID) != claims.PreferredUsername {
		return nil, huma.NewError(http.StatusForbidden, "Forbidden: You can only access your own orders")
	}

	resp.Body = order

	products_url := os.Getenv("PRODUCTS_URL")
	url := fmt.Sprintf("%s/products/%d/orders", products_url, order.ID)
	r, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch products: %v\n", err)
		return resp, nil
	}
	defer r.Body.Close()

	if r.StatusCode == http.StatusOK {
		var productsResp dto.ProductsOutputBody
		if err := json.NewDecoder(r.Body).Decode(&productsResp); err != nil {
			fmt.Printf("Failed to decode products response: %v\n", err)
		} else {
			resp.Body.Products = productsResp.Products
		}
	} else {
		fmt.Printf("Products API returned status %d\n", r.StatusCode)
	}

	return resp, nil
}

func GetOrdersByIdCustomer(ctx context.Context, db *gorm.DB, id uint) (*dto.OrdersOutput, error) {
	
	claims, ok := auth.GetClaims(ctx)
	if !ok {
		return nil, huma.NewError(http.StatusUnauthorized, "Unauthorized")
	}

	// Only admin can access this endpoint or user accessing their own orders
	if !isAdmin(claims) && claims.PreferredUsername != fmt.Sprintf("%d", id) {
		return nil, huma.NewError(http.StatusForbidden, "Forbidden: You can only access your own orders")
	}
	
	resp := &dto.OrdersOutput{}

	var orders []models.Order
	if err := db.Where("customer_id = ?", id).Find(&orders).Error; err != nil {
		return nil, err
	}

	productsURL := os.Getenv("PRODUCTS_URL")
	client := &http.Client{Timeout: 5 * time.Second}
	var wg sync.WaitGroup

	for i := range orders {
		wg.Add(1)
		go func(order *models.Order) {
			defer wg.Done()

			url := fmt.Sprintf("%s/products/%d/orders", productsURL, order.ID)
			r, err := client.Get(url)
			if err != nil {
				fmt.Printf("Failed to fetch products for order %d: %v\n", order.ID, err)
				return
			}
			defer r.Body.Close()

			if r.StatusCode == http.StatusOK {
				var productsResp dto.ProductsOutputBody
				if err := json.NewDecoder(r.Body).Decode(&productsResp); err != nil {
					fmt.Printf("Failed to decode products for order %d: %v\n", order.ID, err)
					return
				}
				order.Products = productsResp.Products
			} else {
				fmt.Printf("Products API for order %d returned status %d\n", order.ID, r.StatusCode)
			}
		}(&orders[i])
	}

	wg.Wait()

	resp.Body.Orders = orders

	return resp, nil
}

// ----------------------
// Register routes with Huma
// ----------------------

func RegisterOrdersRoutes(api huma.API, dbConn *gorm.DB, ch *amqp.Channel) {

	huma.Register(api, huma.Operation{
		OperationID: "get-orders",
		Summary:     "Get all orders",
		Method:      http.MethodGet,
		Path:        "/orders",
		Tags:        []string{"orders"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, func(ctx context.Context, input *struct{}) (*dto.OrdersOutput, error) {
		return GetOrders(ctx, dbConn)
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-order",
		Summary:     "Get a order",
		Method:      http.MethodGet,
		Path:        "/orders/{id}",
		Tags:        []string{"orders"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
	}) (*dto.OrderOutput, error) {
		return GetOrder(ctx, dbConn, input.Id)
	})

	huma.Register(api, huma.Operation{
		OperationID:   "get-customer-orders",
		Summary:       "Get all orders for a customer",
		Method:        http.MethodGet,
		DefaultStatus: http.StatusOK,
		Path:          "/orders/{customerId}/customers",
		Tags:          []string{"orders"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, func(ctx context.Context, input *dto.CustomerOrdersInput) (*dto.OrdersOutput, error) {
		return GetOrdersByIdCustomer(ctx, dbConn, input.CustomerID)
	})

	huma.Register(api, huma.Operation{
		OperationID:   "create-order",
		Summary:       "Create an order",
		Method:        http.MethodPost,
		DefaultStatus: http.StatusCreated,
		Path:          "/orders",
		Tags:          []string{"orders"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, func(ctx context.Context, input *dto.OrderCreateInput) (*dto.OrderOutput, error) {
		
		claims, ok := auth.GetClaims(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "Unauthorized")
		}

		// Users can create orders, but only for themselves
		// Override the CustomerID with the authenticated user's ID
		var customerID uint
		if isAdmin(claims) && input.Body.CustomerID != 0 {
			customerID = input.Body.CustomerID
		} else {
			// Parse PreferredUsername as uint
			_, parseErr := fmt.Sscanf(claims.PreferredUsername, "%d", &customerID)
			if parseErr != nil {
				customerID = input.Body.CustomerID
			}
		}
		
		resp := &dto.OrderOutput{}

		order := models.Order{
			CustomerID: input.Body.CustomerID,
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

		var orderProducts []localModels.OrderProduct

		for _, productID := range input.Body.ProductIDs {
			orderProducts = append(orderProducts, localModels.OrderProduct{
				OrderID:   order.ID,
				ProductID: productID,
			})
		}

		if err := dbConn.Create(&orderProducts).Error; err != nil {
			fmt.Printf("Failed to create OrderProduct records: %v\n", err)
		}

		// Prepare response
		resp.Body = order

		// Publish order created event
		var simplifiedOrder = events.SimplifiedOrder{
			OrderID:    order.ID,
			CustomerID: input.Body.CustomerID,
			ProductIDs: input.Body.ProductIDs,
		}
		if err := rabbitmq.PublishOrderEvent(ch, events.OrderCreated, simplifiedOrder); err != nil {
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
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
		dto.OrderCreateInput
	}) (*dto.OrderOutput, error) {
		
		claims, ok := auth.GetClaims(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "Unauthorized")
		}

		// Only admin can update orders
		if !isAdmin(claims) {
			return nil, huma.NewError(http.StatusForbidden, "Forbidden: Only admins can update orders")
		}

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
		}

		results = dbConn.Model(&order).Updates(updates)
		if results.Error != nil {
			return nil, results.Error
		}

		// Get updated order from DB to ensure all fields are correct
		dbConn.First(&order, order.ID)
		resp.Body = order

		// Publish order updated event
		var simplifiedOrder = events.SimplifiedOrder{
			OrderID:    order.ID,
			CustomerID: order.CustomerID,
		}
		err := rabbitmq.PublishOrderEvent(ch, events.OrderUpdated, simplifiedOrder)
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
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, func(ctx context.Context, input *struct {
		Id uint `path:"id"`
	}) (*struct{}, error) {
		
		claims, ok := auth.GetClaims(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "Unauthorized")
		}

		// Only admin can delete orders
		if !isAdmin(claims) {
			return nil, huma.NewError(http.StatusForbidden, "Forbidden: Only admins can delete orders")
		}
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
			var simplifiedOrder = events.SimplifiedOrder{
				OrderID:    order.ID,
				CustomerID: order.CustomerID,
			}
			err := rabbitmq.PublishOrderEvent(ch, events.OrderDeleted, simplifiedOrder)
			if err != nil {
				// Log the error but don't fail the request
				// The order was already deleted from the database
			}

			return resp, nil
		}

		return nil, results.Error
	})
}
