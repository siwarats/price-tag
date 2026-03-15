package shared

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MONGODB_URI          string
	LOTUSS_SCRAPER_URL   string
	SKIP_EXISTING_IMAGES bool
	SCRAPER_LOTUSS_PORT  string
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
		SCRAPER_LOTUSS_PORT:  os.Getenv("SCRAPER_LOTUSS_PORT"),
	}

	log.Printf("Config: %+v\n", cfg)

	return cfg
}
