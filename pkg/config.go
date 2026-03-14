package pkg

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MONGODB_URI string
	LOTUSS_SCRAPER_URL string
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfg := &Config{
		MONGODB_URI: os.Getenv("MONGODB_URI"),
		LOTUSS_SCRAPER_URL: os.Getenv("LOTUSS_SCRAPER_URL"),
	}

	log.Printf("Config: %+v\n", cfg)

	return cfg
}
