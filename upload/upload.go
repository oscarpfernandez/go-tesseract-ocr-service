package upload

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
	"github.com/oscarpfernandez/go-tesseract-ocr-service/schema"
	"github.com/oscarpfernandez/go-tesseract-ocr-service/wrappers"
)

const NUMBER_PARALELL_ROUTINES = 4

var throttle = make(chan int, NUMBER_PARALELL_ROUTINES)

func GuiUploadPDF(w http.ResponseWriter, req *http.Request) {
	log.Info("Request to upload image service via GUI")

	microPage := `
		<html>
			<title>Hackathon Tesseract Web Service</title>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
			<body>
				<h2>Hackathon Tesseract Web Service</h2>
				<h4>PDF File Submission</h4>
				</pre>	
					<form action="/api/upload/pdf" method="post" enctype="multipart/form-data">
						<input type="file" name="the_file" />
						<input type="submit" value="Submit PDF" />
				</form>
				<pre class="prettyprint">
				<div id="result"></div>
			</body>
		</html>`

	_, _ = fmt.Fprintf(w, microPage)
}

func GuiUploadImage(w http.ResponseWriter, req *http.Request) {
	log.Info("Request to upload image service via GUI")

	microPage := `
		<html>
			<title>Hackathon Tesseract Web Service</title>
			<script src="//ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.min.js"></script>
			<body>
				<h2>Hackathon Tesseract Web Service</h2>
				<h4>JPG Image File Submission</h4>
				</pre>
					<form action="/api/upload/img" method="post" enctype="multipart/form-data">
						<input type="file" name="the_file" />
						<input type="submit" value="Submit JPG" />
				</form>
				<pre class="prettyprint">
				<div id="result"></div>
			</body>
		</html>`

	_, _ = fmt.Fprintf(w, microPage)
}

func UploadImage(w http.ResponseWriter, req *http.Request) {
	log.Info("Request to upload image service")

	var (
		err        error
		submission schema.SubmissionDetails
	)

	if !validateInput(w, req, &submission) {
		log.WithField("submissions", submission).Error("Invalid submission")
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	var tempPath string
	var numberOfPages int
	var txtsOutputPath string

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			submission.FileName = hdr.Filename
			// open uploaded
			var infile multipart.File
			if infile, err = hdr.Open(); err != nil {
				log.WithField("imgFilename", hdr.Filename).WithError(err).Error("Error uploading image file")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}
			// open destination
			var outfile *os.File

			// Save the file into the docker container disk,
			generatedUUID, err := uuid.NewV4()
			if err != nil {
				log.WithError(err).Error("Error creating UUID")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			submission.UUID = generatedUUID.String()

			tempPath = path.Join(os.Getenv("UPLOADED_FILES_DIR"), generatedUUID.String())
			log.WithFields(log.Fields{
				"tmpDir":   tempPath,
				"fileName": hdr.Filename,
			}).Info("Storing submitted Image")

			if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"tempPath": tempPath,
				}).Error("Unable to write temporary folder")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			outfile, err = os.Create(filepath.Join(tempPath, schema.DocumentImageName))
			if err != nil {
				log.WithError(err).Error("Error creating temporary upload file")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}
			defer outfile.Close()

			// 32K buffer copy
			if _, err = io.Copy(outfile, infile); err != nil {
				log.WithError(err).Error("Error while copying file")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			log.WithField("MaxConcurrency", NUMBER_PARALELL_ROUTINES).Info("Launching main Tesseract text extraction worker")
			txtsOutputPath = path.Join(tempPath, schema.TextFolderName)
			if err := os.MkdirAll(txtsOutputPath, os.ModePerm); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"txtsOutputPath": txtsOutputPath,
				}).Error("Unable to write text output folder")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			var wg sync.WaitGroup
			numberOfPages = processParalellOCR(tempPath, "jpg", txtsOutputPath, &wg, throttle)
		}
	}

	submission.NumberOfPages = numberOfPages
	submission.Pages = generatePageDetails(txtsOutputPath)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(submission); err != nil {
		log.WithError(err).Error("Error marshalling submission JSON")
	}
}

func UploadPDF(w http.ResponseWriter, req *http.Request) {
	log.Info("Request to upload pdf service")

	var (
		err        error
		submission schema.SubmissionDetails
	)

	if !validateInput(w, req, &submission) {
		log.WithField("submissions", submission).Error("Invalid submission")
		http.Error(w, "Unable to process request", http.StatusBadRequest)
		return
	}

	var tempPath string
	var numberOfPages int
	var txtsOutputPath string

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			submission.FileName = hdr.Filename
			// open uploaded
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				log.WithField("PDFFilename", hdr.Filename).WithError(err).Error("Error uploading PDF file")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}
			// open destination
			var outfile *os.File

			// Save the file into the docker container disk,
			// Save the file into the docker container disk,
			generatedUUID, err := uuid.NewV4()
			if err != nil {
				log.WithError(err).Error("Error creating UUID")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			submission.UUID = generatedUUID.String()

			tempPath = path.Join(os.Getenv("UPLOADED_FILES_DIR"), generatedUUID.String())
			log.WithFields(log.Fields{
				"tmpDir":   tempPath,
				"fileName": hdr.Filename,
			}).Info("Storing submitted PDF")

			if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
				log.WithError(err).Error("Error creating temporary directory")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			if outfile, err = os.Create(filepath.Join(tempPath, schema.DocumentFileName)); nil != err {
				log.WithError(err).Error("Error creating temporary upload file")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}
			defer outfile.Close()

			// 32K buffer copy
			if _, err = io.Copy(outfile, infile); nil != err {
				log.WithError(err).Error("Error while copying file")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			// Generates Images from the PDF
			imagesOutputPath := path.Join(tempPath, schema.ImagesFolderName)
			if err := os.MkdirAll(imagesOutputPath, os.ModePerm); err != nil {
				log.WithError(err).Error("Error creating images output directory")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			pdfFilePath := path.Join(tempPath, schema.DocumentFileName)
			if err := wrappers.ExtracPdfToImagesFromPDF(pdfFilePath, imagesOutputPath); err != nil {
				log.WithField("pdfFilePath", pdfFilePath).WithError(err).Error("Unable to extract images from PDF")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			var wg sync.WaitGroup
			log.WithField("MaxConcurrency", NUMBER_PARALELL_ROUTINES).Info("Launching main Tesseract text extraction worker")
			txtsOutputPath = path.Join(tempPath, schema.TextFolderName)
			if err := os.MkdirAll(txtsOutputPath, os.ModePerm); err != nil {
				log.WithError(err).Error("Error creating texts output directory")
				http.Error(w, "Unable to process request", http.StatusInternalServerError)
				return
			}

			numberOfPages = processParalellOCR(imagesOutputPath, "jpg", txtsOutputPath, &wg, throttle)

		}
	}
	submission.NumberOfPages = numberOfPages
	submission.Pages = generatePageDetails(txtsOutputPath)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(submission); err != nil {
		log.WithError(err).Error("Error marshalling submission JSON")
	}

}

func processParalellOCR(imagesDirectoryPath string, imageExtension string, textOutPutDirectory string, wg *sync.WaitGroup, throttle chan int) int {
	imageFilesList, _ := ioutil.ReadDir(imagesDirectoryPath)

	numberFiles := 0

	for _, f := range imageFilesList {
		if !strings.HasSuffix(f.Name(), imageExtension) || f.IsDir() {
			continue
		}
		imagePath := path.Join(imagesDirectoryPath, f.Name())
		throttle <- 1 // whatever number
		wg.Add(1)
		log.WithFields(log.Fields{
			"imagesDirectoryPath": imagesDirectoryPath,
			"imageExtension":      imageExtension,
			"textOutPutDirectory": textOutPutDirectory,
		})
		go wrappers.ExtractPlainTextFromImage(imagePath, "eng", textOutPutDirectory, f.Name(), wg, throttle)

		numberFiles++
	}
	wg.Wait()

	return numberFiles
}

func generatePageDetails(textsDirectory string) []schema.PageDetails {
	txtsFilesList, _ := ioutil.ReadDir(textsDirectory)

	pages := make([]schema.PageDetails, len(txtsFilesList))

	pageNumber := 0

	for _, txt := range txtsFilesList {
		txtPath := path.Join(textsDirectory, txt.Name())
		data, err := ioutil.ReadFile(txtPath)

		if err != nil {
			log.WithError(err).Error("Cannot read txt file")
		}

		page := schema.PageDetails{
			PageNumber: pageNumber + 1,
			Text:       string(data),
		}
		pages[pageNumber] = page
		pageNumber++
	}

	return pages

}

func validateInput(w http.ResponseWriter, req *http.Request, submission *schema.SubmissionDetails) bool {
	// Need to call ParseMultipartForm first so we can check if a file is contained
	// parameter for max memory taken from https://golang.org/src/net/http/request.go
	// Note that this is 32mb, and is probably why 40mb files are failing
	_ = req.ParseMultipartForm(32 << 20)

	if req.MultipartForm == nil || len(req.MultipartForm.File) == 0 {
		log.Error("No file passed in")
		return false
	}

	var maxSizeBits int64
	maxSizeBits = (1 << 20) * schema.MaxSizeMB

	if err := req.ParseMultipartForm(maxSizeBits); nil != err {
		log.Error("File exceeds maximum size")
		return false
	}
	return true
}
