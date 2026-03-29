package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	apiKey, _ := os.LookupEnv("API_KEY")

	viper.Set("DATABASE_URL", "postgres://"+dbHost+":"+dbPort)
	viper.Set("SECRET_KEY", apiKey)

	fmt.Println("Configured:", viper.Get("DATABASE_URL"))
}
