package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/oscarpfernandez/go-tesseract-ocr-service/upload"
	"github.com/rs/cors"
)

var log = logrus.New()

func init() {
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel
}

//go:generate go-bindata-assetfs ../../view/...
func main() {
	logrus.Print("Tesseract Rest Service")

	handlers := http.NewServeMux()
	handlers.HandleFunc("/api/upload/pdf", upload.UploadPDF)
	handlers.HandleFunc("/api/upload/img", upload.UploadImage)
	handlers.HandleFunc("/web/pdf", upload.GuiUploadPDF)
	handlers.HandleFunc("/web/img", upload.GuiUploadImage)

	// Wrapping the API handler in CORS default behaviors
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
		AllowedMethods: []string{"GET", "POST"},
	})
	h := c.Handler(handlers)

	// Assign the returned mux to the default http root handler
	http.Handle("/", h)

	// Setup a go-bindata-assetfs file server on the /view route
	http.Handle("/view/", http.StripPrefix("/view/", http.FileServer(assetFS())))

	// Start server
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		logrus.Fatal("Error attempting to ListenAndServe: ", err)
	}
}
