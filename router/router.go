package router

import (
	"github.com/oscarpfernandez/go-tesseract-ocr-service/upload"
	"github.com/bmizerany/pat"
)

// Handlers API handlers
func Handlers() *pat.PatternServeMux {
	m := pat.New()
	m.Post("/api/upload/pdf", upload.UploadPDF())
	m.Post("/api/upload/img", upload.UploadImage())
	m.Get("/web/pdf", upload.GuiUploadPDF())
	m.Get("/web/img", upload.GuiUploadImage())

	return m
}
