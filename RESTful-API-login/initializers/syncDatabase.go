package initializers

import "example/RESTful-API-login/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}
