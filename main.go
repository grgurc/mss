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
const serverAddress string = "http://localhost:8080"

type InitialUploadData struct {
	AudioPath string
	ImagePath string
}

type SeparatedData struct {
	SourceType string // type of source (harmonic, percussive, vocals...)
	AudioPath  string
	ImagePath  string
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
	afterSeparate := template.Must(template.ParseFiles("./templates/separation_result.html"))

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
			AudioPath: fmt.Sprintf("%v/%v", serverAddress, "uploads/original.wav"),
			ImagePath: fmt.Sprintf("%v/%v", serverAddress, "uploads/original_spectrogram.png"),
		}

		afterUpload.Execute(w, d)
	})

	s.Get("/separate-median", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /separate-median")

		cmd := exec.Command(pythonPath, "./scripts/separate_median.py")
		output, err := cmd.CombinedOutput()
		log.Println(string(output))
		log.Println(err)

		cmd = exec.Command(pythonPath, "./scripts/spectrogram.py", "./uploads/median_harmonic.wav")
		output, err = cmd.CombinedOutput()
		log.Println(string(output))
		log.Println(err)

		cmd = exec.Command(pythonPath, "./scripts/spectrogram.py", "./uploads/median_percussive.wav")
		output, err = cmd.CombinedOutput()
		log.Println(string(output))
		log.Println(err)

		d := []SeparatedData{
			{
				SourceType: "Harmonic",
				AudioPath:  fmt.Sprintf("%v/%v", serverAddress, "uploads/median_harmonic.wav"),
				ImagePath:  fmt.Sprintf("%v/%v", serverAddress, "uploads/median_harmonic_spectrogram.png"),
			},
			{
				SourceType: "Percussive",
				AudioPath:  fmt.Sprintf("%v/%v", serverAddress, "uploads/median_percussive.wav"),
				ImagePath:  fmt.Sprintf("%v/%v", serverAddress, "uploads/median_percussive_spectrogram.png"),
			},
		}

		log.Println(d)

		afterSeparate.Execute(w, d)
	})

	fmt.Println("Starting server on localhost:8080")
	if err := http.ListenAndServe(":8080", s); err != nil {
		log.Fatal(err)
	}
}
