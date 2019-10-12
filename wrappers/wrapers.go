package wrappers

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/oscarpfernandez/go-tesseract-ocr-service/schema"

	log "github.com/Sirupsen/logrus"
	"github.com/otiai10/gosseract"
)

//ExtracPdfToImagesFromPDF extracts Images from the PDF file and output an image per page.
func ExtracPdfToImagesFromPDF(pdfFullPath, outputDirectory string) error {
	log.WithFields(log.Fields{
		"pdfFullPath":     pdfFullPath,
		"outputDirectory": outputDirectory,
	}).Info("Extracting Images from PDF via Ghostscript")

	err := os.Chdir(outputDirectory)
	if err != nil {
		log.WithField("basePath", pdfFullPath).WithError(err).Error("Chdir dir failed")
	}

	cmdArgs := []string{"-dNOPAUSE", "-dBATCH", "-sDEVICE=jpeg", "-r300", "-sOutputFile=p%03d.jpg", pdfFullPath}

	cmd := exec.Command("gs", cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.WithError(err).Error("Error creating StdoutPipe for Cmd")
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Printf("Ghosscript output | %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.WithError(err).Error("Error starting Cmd")
		return err
	}

	err = cmd.Wait()
	if err != nil {
		log.WithError(err).Error("Error waiting for Cmd")
		return err
	}

	return nil
}

//ExtractPlainTextFromImage given a images file, Tesseract OCR generates a plain text file with the detected text.
func ExtractPlainTextFromImage(imageFullPath, languages, outputDirectory, textFilePrefix string, wg *sync.WaitGroup, throttle chan int) {
	defer wg.Done()

	outText := gosseract.Must(gosseract.Params{
		Src:       imageFullPath,
		Languages: languages, //eng+heb
	})

	textFilePath := filepath.Join(outputDirectory, fmt.Sprintf("%s_%s", textFilePrefix, schema.TextFileName))
	outfile, err := os.Create(textFilePath)
	if err != nil {
		log.WithError(err).Error("Error creating text file")
		return
	}
	defer outfile.Close()

	log.WithFields(log.Fields{
		"imageFullPath":   imageFullPath,
		"outputDirectory": outputDirectory,
		"languages":       languages,
		"textFilePath":    textFilePath,
	}).Info("Processed OCR Tesseract Instance")

	sanitizedTxt := strings.Replace(outText, "\n", " ", -1)
	outfile.WriteString(sanitizedTxt)

	<-throttle
}
