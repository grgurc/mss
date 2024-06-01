package api

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grgurcz/mss/python"
	sep "github.com/grgurcz/mss/separator"
	"github.com/grgurcz/mss/templating"
	"github.com/grgurcz/mss/util"
)

type Api struct {
	uploadsPath string
	separator   *sep.Separator
	python      *python.Python
	templater   *templating.Templater
}

func NewApi(
	basePath string,
	separator *sep.Separator,
	python *python.Python,
	templater *templating.Templater,
) *Api {
	uploadsPath := path.Join(basePath, "uploads")
	return &Api{
		uploadsPath: uploadsPath,
		separator:   separator,
		python:      python,
		templater:   templater,
	}
}

func (a *Api) Index(w http.ResponseWriter, r *http.Request) {
	a.templater.Index(w)
}

func (a *Api) Separate(w http.ResponseWriter, r *http.Request) {
	method := chi.URLParam(r, "method")
	log.Printf("GET /separate/%v", method)

	err := a.separator.Separate(method)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := []templating.AudioData{}
	dirPath := path.Join(a.uploadsPath, method)

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && util.IsWav(d) {
			sourceType := strings.Split(d.Name(), ".")[0]
			audioName := sourceType + ".wav"
			imageName := sourceType + "_spectrogram.png"

			data = append(data, templating.AudioData{
				SourceType: sourceType,
				AudioPath:  util.URL(dirPath, audioName),
				ImagePath:  util.URL(dirPath, imageName),
			})
		}

		return nil
	})

	fmt.Println(data)

	a.templater.AfterSeparate(w, data)
}

func (a *Api) HandleUpload(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /upload")

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()

	originalPath, err := a.saveUploadedFile(f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = a.python.Spectrogram(originalPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := templating.UploadData{
		Methods:   []string{"median", "repet", "unmix"},
		AudioPath: util.URL("uploads/original.wav"),
		ImagePath: util.URL("uploads/original_spectrogram.png"),
	}

	a.templater.AfterUpload(w, data)
}

func (a *Api) saveUploadedFile(f multipart.File) (string, error) {
	p := path.Join(a.uploadsPath, "original.wav")
	dst, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, f)
	if err != nil {
		return "", err
	}

	return p, nil
}
