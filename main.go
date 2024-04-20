package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-chi/chi/v5"
)

const pythonPath string = "./env/bin/python3"

type InitialUploadData struct {
	AudioPath string
	ImagePath string
}

func saveUploadedFile(f multipart.File) error {
	err := os.Mkdir("./uploads", os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	dst, err := os.Create("./uploads/original.wav")
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, f)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	index := template.Must(template.ParseFiles("./templates/index.html"))
	afterUpload := template.Must(template.ParseFiles("./templates/after_upload.html"))

	s := chi.NewRouter()

	s.Get("/uploads/{file}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /uploads/%s", chi.URLParam(r, "file"))
		http.ServeFile(w, r, "./uploads/"+chi.URLParam(r, "file"))
	})

	s.Get("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /")
		index.Execute(w, nil)
	})

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
		output, err := cmd.CombinedOutput()
		log.Println(string(output))
		log.Println(err)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		d := InitialUploadData{
			AudioPath: "http://localhost:8080/uploads/original.wav",
			ImagePath: "http://localhost:8080/uploads/original_spectrogram.png",
		}

		afterUpload.Execute(w, d)
		return
		// now here we can return the spectrogram, not sure how though xd lol
		// i think we need the address of the original request (aka the base url of the server)
		// and then we are going to need something else which i cant think of right now
		w.Write([]byte("<img src=\"http://localhost:8080/uploads/original_spectrogram.png\">"))
		// afterUpload.Execute(w, nil) // TODO
	})

	s.Get("/separate-median", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /separate-median")
		cmd := exec.Command(pythonPath, "./scripts/separate_median.py", "./uploads/original.wav")
		output, err := cmd.CombinedOutput()
		log.Println(string(output))
		log.Println(err)

		return
		// afterSeparate.Execute(w, nil)
	})

	fmt.Println("Starting server on localhost:8080")
	if err := http.ListenAndServe(":8080", s); err != nil {
		log.Fatal(err)
	}
}
