package main

import (
	"net/http"

	"github.com/oscarpfernandez/go-tesseract-ocr-service/router"
	"github.com/Sirupsen/logrus"
	"github.com/rs/cors"
)

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel
}

//go:generate go-bindata-assetfs view/...
func main() {
	logrus.Print("Tesseract Rest Service")

	handler := getAPIHandlers()

	// Assign the returned mux to the default http root handler
	http.Handle("/", handler)

	// Setup a go-bindata-assetfs file server on the /view route
	http.Handle("/view/", http.StripPrefix("/view/", http.FileServer(assetFS())))

	// Start server
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		logrus.Fatal("Error attempting to ListenAndServe: ", err)
	}
}

func getAPIHandlers() http.Handler {
	// Wrapping the API handler in CORS default behaviors
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
		AllowedMethods: []string{"GET", "POST"},
	})
	handler := c.Handler(router.Handlers())
	return handler
}
