package lotuss

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (l *lotuss) runImages() {
	ctx := context.Background()
	cur, err := l.db.Collection(PRODUCT_COLLECTION).Find(ctx, bson.M{},
		options.Find().SetBatchSize(100).SetProjection(bson.M{"thumbnail": 1}))
	if err != nil {
		log.Fatalf("failed to query products: %v", err)
	}
	defer cur.Close(ctx)

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for cur.Next(ctx) {
		var doc map[string]interface{}
		if err := cur.Decode(&doc); err != nil {
			continue
		}

		thumbnail, ok := doc["thumbnail"].(map[string]interface{})
		if !ok {
			continue
		}
		imgURL, _ := thumbnail["url"].(string)
		if imgURL == "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() { <-sem }()
			downloadImage(u, l.skipExistingImages)
		}(imgURL)
	}

	wg.Wait()
	log.Println("done downloading images")
}

func downloadImage(rawURL string, skipExisting bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("parse url %s: %v", rawURL, err)
		return
	}

	localPath := filepath.Join("storage", u.Path)

	if skipExisting {
		if _, err := os.Stat(localPath); err == nil {
			return
		}
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
		log.Printf("mkdir %s: %v", filepath.Dir(localPath), err)
		return
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		log.Printf("download %s: %v", rawURL, err)
		return
	}
	defer resp.Body.Close()

	f, err := os.Create(localPath)
	if err != nil {
		log.Printf("create %s: %v", localPath, err)
		return
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		log.Printf("write %s: %v", localPath, err)
	}
	log.Printf("saved %s", localPath)
}
