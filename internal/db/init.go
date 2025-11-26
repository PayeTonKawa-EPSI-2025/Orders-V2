package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/PayeTonKawa-EPSI-2025/Common-V2/models"

	localModels "github.com/PayeTonKawa-EPSI-2025/Orders-V2/internal/models"
)

func Init() *gorm.DB {
	dsn := os.Getenv("DATABASE_DSN")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	db.AutoMigrate(&models.Order{}, &localModels.Customer{}, &localModels.Product{}, &localModels.CustomerOrder{}, &localModels.OrderProduct{})

	return db
}
