package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/grgurcz/mss/api"
	"github.com/grgurcz/mss/python"
	"github.com/grgurcz/mss/separator"
	"github.com/grgurcz/mss/templating"
)

const serverAddress string = "http://localhost:8080"

func initFolders() (string, error) {
	basePath := "."
	err := os.MkdirAll(path.Join(basePath, "uploads"), os.ModePerm)
	if err != nil {
		return "", err
	}

	return basePath, nil
}

func main() {
	basePath, err := initFolders()
	if err != nil {
		panic(err)
	}

	python := python.NewPython(basePath)

	fmt.Println("python")
	fmt.Println(python)

	separator := separator.NewSeparator(
		basePath,
		serverAddress,
		python,
	)

	templater := templating.NewTemplater(basePath)

	api := api.NewApi(
		basePath,
		separator,
		python,
		templater,
	)

	s := chi.NewRouter()

	// TODO: put these in Api as well i guess
	s.Get("/uploads/{file}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /uploads/%s", chi.URLParam(r, "file"))
		http.ServeFile(w, r, fmt.Sprintf("./uploads/%s", chi.URLParam(r, "file")))
	})

	s.Get("/uploads/{folder}/{file}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /uploads/%s/%s", chi.URLParam(r, "folder"), chi.URLParam(r, "file"))
		http.ServeFile(w, r, fmt.Sprintf("./uploads/%s/%s", chi.URLParam(r, "folder"), chi.URLParam(r, "file")))
	})

	s.Get("/", api.Index)
	s.Post("/upload", api.HandleUpload)
	s.Get("/separate/{method}", api.Separate)

	/*
		s.Post("/upload", func(w http.ResponseWriter, r *http.Request) {
			log.Println("POST /upload")
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			f, _, err := r.FormFile("file")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer f.Close()

			err = saveUploadedFile(f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			cmd := exec.Command(pythonPath, "./scripts/spectrogram.py", "./uploads/original.wav")
			_, err = cmd.CombinedOutput()
			if err != nil {
				log.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			d := InitialUploadData{
				AudioPath: fmt.Sprintf("%v/%v", serverAddress, "uploads/original.wav"),
				ImagePath: fmt.Sprintf("%v/%v", serverAddress, "uploads/original_spectrogram.png"),
			}

			afterUpload.Execute(w, d)
		})
	*/

	fmt.Println("Starting server on localhost:8080")
	if err := http.ListenAndServe(":8080", s); err != nil {
		log.Fatal(err)
	}
}
