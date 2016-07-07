package upload

import (
	"net/http"
	"sync"

	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path"

	"io/ioutil"

	"strings"

	"fmt"

	"github.com/oscarpfernandez/go-tesseract-ocr-service/schema"
	"github.com/oscarpfernandez/go-tesseract-ocr-service/wrappers"
	log "github.com/Sirupsen/logrus"
	"github.com/nu7hatch/gouuid"
)

const NUMBER_PARALELL_ROUTINES = 4

var throttle = make(chan int, NUMBER_PARALELL_ROUTINES)

func GuiUploadPDF() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		fmt.Fprintf(w, microPage)
	}

}

func GuiUploadImage() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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

		fmt.Fprintf(w, microPage)
	}

}

func UploadImage() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Info("Request to upload image service")

		w.Header().Set("Content-Type", "application/json")
		var (
			status     int
			err        error
			submission schema.SubmissionDetails
		)
		// Runs at the end for error handline - check util for errors?
		defer func() {
			if nil != err {
				log.WithField("function", "uploadImage").WithError(err).Error("Error uploading image file")
				http.Error(w, err.Error(), status)
			}
		}()

		if validateInput(w, req, &submission) == false {
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
					status = http.StatusInternalServerError
					return
				}
				// open destination
				var outfile *os.File

				// Save the file into the docker container disk,
				generatedUUID, err := uuid.NewV4()

				submission.UUID = generatedUUID.String()

				if err != nil {
					status = http.StatusInternalServerError
					log.WithError(err).Error("Error creating UUID")
					return
				}

				tempPath = path.Join(os.Getenv("UPLOADED_FILES_DIR"), generatedUUID.String())
				log.WithFields(log.Fields{
					"tmpDir":   tempPath,
					"fileName": hdr.Filename,
				}).Info("Storing submitted Image")

				os.MkdirAll(tempPath, os.ModePerm)

				if outfile, err = os.Create(tempPath + "/" + schema.DOCUMENT_IMAGE); nil != err {
					status = http.StatusInternalServerError
					log.WithError(err).Error("Error creating temporary upload file")
					return
				}
				defer outfile.Close()

				// 32K buffer copy
				if _, err = io.Copy(outfile, infile); nil != err {
					status = http.StatusInternalServerError
					return
				}

				var wg sync.WaitGroup
				log.WithField("MaxConcurrency", NUMBER_PARALELL_ROUTINES).Info("Launching main Tesseract text extraction worker")
				txtsOutputPath = path.Join(tempPath, schema.TEXT_FOLDER)
				os.MkdirAll(txtsOutputPath, os.ModePerm)
				numberOfPages = processParalellOCR(tempPath, "jpg", txtsOutputPath, &wg, throttle)

			}
		}
		submission.NumberOfPages = numberOfPages
		submission.Pages = generatePageDetails(txtsOutputPath)

		b, err := json.MarshalIndent(submission, "", "\t")
		if err != nil {
			log.WithError(err).Error("Error marshalling submission JSON")
		}
		w.Write(b)
	}
}

func UploadPDF() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Info("Request to upload pdf service")

		w.Header().Set("Content-Type", "application/json")
		var (
			status     int
			err        error
			submission schema.SubmissionDetails
		)
		// Runs at the end for error handline - check util for errors?
		defer func() {
			if nil != err {
				log.WithField("function", "uploadPDF").WithError(err).Error("Error uploading pdf file")
				http.Error(w, err.Error(), status)
			}
		}()

		if validateInput(w, req, &submission) == false {
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
					status = http.StatusInternalServerError
					return
				}
				// open destination
				var outfile *os.File

				// Save the file into the docker container disk,
				generatedUUID, err := uuid.NewV4()

				submission.UUID = generatedUUID.String()

				if err != nil {
					status = http.StatusInternalServerError
					log.WithError(err).Error("Error creating UUID")
					return
				}

				tempPath = path.Join(os.Getenv("UPLOADED_FILES_DIR"), generatedUUID.String())
				log.WithFields(log.Fields{
					"tmpDir":   tempPath,
					"fileName": hdr.Filename,
				}).Info("Storing submitted PDF")

				os.MkdirAll(tempPath, os.ModePerm)

				if outfile, err = os.Create(tempPath + "/" + schema.DOCUMENT_FILE); nil != err {
					status = http.StatusInternalServerError
					log.WithError(err).Error("Error creating temporary upload file")
					return
				}
				defer outfile.Close()

				// 32K buffer copy
				if _, err = io.Copy(outfile, infile); nil != err {
					status = http.StatusInternalServerError
					return
				}

				// Generates Images from the PDF
				imagesOutputPath := path.Join(tempPath, schema.IMAGES_FOLDER)
				os.MkdirAll(imagesOutputPath, os.ModePerm)
				pdfFilePath := path.Join(tempPath, schema.DOCUMENT_FILE)
				wrappers.ExtracPdfToImagesFromPDF(pdfFilePath, imagesOutputPath)

				var wg sync.WaitGroup
				log.WithField("MaxConcurrency", NUMBER_PARALELL_ROUTINES).Info("Launching main Tesseract text extraction worker")
				txtsOutputPath = path.Join(tempPath, schema.TEXT_FOLDER)
				os.MkdirAll(txtsOutputPath, os.ModePerm)
				numberOfPages = processParalellOCR(imagesOutputPath, "jpg", txtsOutputPath, &wg, throttle)

			}
		}
		submission.NumberOfPages = numberOfPages
		submission.Pages = generatePageDetails(txtsOutputPath)

		b, err := json.MarshalIndent(submission, "", "\t")
		if err != nil {
			log.WithError(err).Error("Error marshalling submission JSON")
		}
		w.Write(b)
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
	maxSizeBits = (1 << 20) * schema.MAXSIZEMB

	if err := req.ParseMultipartForm(maxSizeBits); nil != err {
		log.Error("File exceeds maximum size")
		return false
	}
	return true
}
