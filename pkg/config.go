package pkg

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MONGODB_URI          string
	LOTUSS_SCRAPER_URL   string
	SKIP_EXISTING_IMAGES bool
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfg := &Config{
		MONGODB_URI:          os.Getenv("MONGODB_URI"),
		LOTUSS_SCRAPER_URL:   os.Getenv("LOTUSS_SCRAPER_URL"),
		SKIP_EXISTING_IMAGES: os.Getenv("SKIP_EXISTING_IMAGES") == "true",
	}

	log.Printf("Config: %+v\n", cfg)

	return cfg
}
