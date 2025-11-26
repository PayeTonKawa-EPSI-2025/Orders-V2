package operation

import "github.com/PayeTonKawa-EPSI-2025/Common-V2/models"

// Mock function to replace PublishCustomerEvent
var PublishOrderEvent = func(ch any, event string, customer models.Order) error {
	return nil
}
