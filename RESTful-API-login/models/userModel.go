package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	// ID       int    `gorm:"identity(1,1)"`
	Fname    string
	Lname    string
	Phone    string
	Email    string `gorm:"unique"`
	Password string
	Adm      bool `gorm:"default:true"`
}
