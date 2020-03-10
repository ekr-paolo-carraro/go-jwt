package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.POST("/signup", signupHandler)
	router.POST("/login", loginHandler)
	router.GET("/protected", authMiddleware(), protectedEndpointHandler)

	log.Fatal(router.Run(":8080"))

}

func signupHandler(c *gin.Context) {
	c.String(http.StatusOK, "hello signupHandler")
}

func loginHandler(c *gin.Context) {
	c.String(http.StatusOK, "hello loginHandler")
}

func protectedEndpointHandler(c *gin.Context) {
	log.Println("in protectedEndpointHandler")
	c.String(http.StatusOK, "hello protectedEndpointHandler")
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("hey sono man in the middle before")
		c.Next()
		log.Println("hey sono man in the middle after")

	}
}
