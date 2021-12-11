package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog/log"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FileName() == "" { // this is FormData
			data, _ := ioutil.ReadAll(part)
			fmt.Printf("FormData=[%s]\n", string(data))
		} else { // This is FileData
			log.Info().Str("filename", part.FileName()).Msg("保存文件")
			dst, _ := os.Create(filepath.Join(savePath, part.FileName()))
			defer dst.Close()
			io.Copy(dst, part)
		}
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

var savePath string

func main() {
	var port int
	flag.IntVar(&port, "port", 8877, "port")
	flag.StringVar(&savePath, "save", ".", "savePath")
	flag.Parse()
	if exist, _ := PathExists(savePath); !exist {
		os.MkdirAll(savePath, 0644)
	}
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
