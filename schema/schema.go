package schema

const (
	// MaxSizeMB - The maximum allowed file size
	MaxSizeMB         = 40
	DocumentFileName  = "document.pdf"
	DocumentImageName = "image.jpg"
	TextFileName      = "text.txt"
	ImagesFolderName  = "images"
	TextFolderName    = "texts"
)

// PageDetails represents the basic elements of page details.
type PageDetails struct {
	PageNumber int    `json:"page_number"`
	Text       string `json:"text"`
}

// SubmissionDetails represents the element details of a submission.
type SubmissionDetails struct {
	UUID          string        `json:"uuid"`
	FileName      string        `json:"pdf_filename"`
	NumberOfPages int           `json:"number_pages"`
	Pages         []PageDetails `json:"page_details"`
}
