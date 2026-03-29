package main

import (
	"fmt"
	"os"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	appPort := os.Getenv("APP_PORT")
	apiKey := os.Getenv("API_KEY")
	secretKey := os.Getenv("SECRET_KEY")

	fmt.Println(dbURL, jwtSecret, appPort, apiKey, secretKey)
}