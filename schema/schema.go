package schema

// MAXSIZEMB - The maximum allowed file size
const MAXSIZEMB = 40
const DOCUMENT_FILE = "document.pdf"
const DOCUMENT_IMAGE = "image.jpg"
const TEXT_FILE = "text.txt"
const IMAGES_FOLDER = "images"
const TEXT_FOLDER = "texts"

//PageDetails
type PageDetails struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
}

// SubmissionDetails
type SubmissionDetails struct {
	UUID          string        `json:"uuid"`
	FileName      string        `json:"pdf_filename"`
	NumberOfPages int           `json:"number_pages"`
	Pages         []PageDetails `json:"page_details"`
}
