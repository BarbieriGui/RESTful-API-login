package main

//https://www.youtube.com/watch?v=ma7rUS_vW9M&ab_channel=CodingwithRobby

//To create a mod » go mod init

//To use compiler run in Bash;
//compiledaemon --command="./RESTful-API-login"

import (
	"example/RESTful-API-login/controllers"
	"example/RESTful-API-login/initializers"
	"example/RESTful-API-login/middleware"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()

	r.POST("/signup", controllers.Signup)
	r.POST("/login", controllers.Login)
	r.POST("/update", controllers.Update)
	r.GET("/validate", middleware.RequireAuth, controllers.Validate)
	r.GET("/listusers", controllers.ListUsers)
	r.Run()
}
