package main

import (
	"net/http"
	"os"

	"github.com/oscarpfernandez/go-tesseract-ocr-service/handlers"

	"github.com/Sirupsen/logrus"
	"github.com/rs/cors"
)

func init() {
	log := logrus.New()
	log.Formatter = new(logrus.TextFormatter) // default
	log.Level = logrus.DebugLevel
}

func main() {
	logrus.Print("Tesseract Rest Service")

	h := handlers.NewHandlers(os.Getenv("UPLOADED_FILES_DIR"))

	router := http.NewServeMux()
	router.HandleFunc("/api/upload/pdf", h.UploadPDF)
	router.HandleFunc("/api/upload/img", h.UploadImage)
	router.HandleFunc("/web/pdf", h.GuiUploadPDF)
	router.HandleFunc("/web/img", h.GuiUploadImage)

	// Wrapping the API handler in CORS default behaviors
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Origin", "Accept", "Content-Type"},
		AllowedMethods: []string{"GET", "POST"},
	})

	// Start server
	err := http.ListenAndServe(":80", c.Handler(router))
	if err != nil {
		logrus.Fatal("Error attempting to ListenAndServe: ", err)
	}
}
