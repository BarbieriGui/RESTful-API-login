package initializers

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// it was necessary to create a DB, witch was done at elephantsql.com
// to make db available everywhere, we create a global var

var DB *gorm.DB

func ConnectToDb() {
	// postgres://tzlebhzj:QXS1MK3PuteKwNgJMqC_1pjm3EkBKENN@manny.db.elephantsql.com/tzlebhzj » from elephantsql, info used to fill dsn info
	// changed port to default 5432
	// Timezone in dsn deleted
	var err error
	dsn := os.Getenv("DB")
	//originally it was db, err := gorm, which means we were creating the variables db and err (for error)
	//a global and local var's where created, and we changed from declaring to assigin value "DB, err = gorm.." insted of ":="
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to db")
	}
}
