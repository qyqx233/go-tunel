package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	var host, filename string
	flag.StringVar(&host, "h", "host", "upload destination")
	flag.StringVar(&filename, "f", "filename", "filename")
	flag.Parse()
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, _ := bodyWriter.CreateFormFile("f", filepath.Base(filename))

	file, _ := os.Open(filename)
	defer file.Close()

	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, _ := http.Post(host, contentType, bodyBuffer)
	if resp != nil {
		defer resp.Body.Close()
	}

	resp_body, _ := ioutil.ReadAll(resp.Body)

	log.Println(resp.Status)
	log.Println(string(resp_body))
}
