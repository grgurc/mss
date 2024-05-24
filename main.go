package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

const (
	pythonPath  string = "env/bin/python3"
	unmixPath   string = "env/bin/umx"
	uploadsPath string = "uploads"
)

const (
	spectrogramScript string = "./scripts/spectrogram.py"
)

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

func timeFunc(f func(args ...any) error) func(args ...any) error {
	return func(args ...any) error {
		s := time.Now()
		err := f(args)
		e := time.Now()
		fmt.Printf("Operation took: %v seconds\n", e.Sub(s).Seconds())
		return err
	}
}

func saveUploadedFile(f multipart.File) error {
	err := os.MkdirAll("./uploads/median", os.ModePerm)
	if err != nil {
		return err
	}

	err = os.MkdirAll("./uploads/umx", os.ModePerm)
	if err != nil {
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

func isWav(d fs.DirEntry) bool {
	if d.IsDir() {
		return false
	}
	return strings.Split(d.Name(), ".")[1] == "wav"
}

func getSeparatedData(dirName string) []SeparatedData {
	var res []SeparatedData
	dirPath := filepath.Join(uploadsPath, dirName) // uploads/{dirName}

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && isWav(d) {
			sourceType := strings.Split(d.Name(), ".")[0]

			res = append(res, SeparatedData{
				SourceType: sourceType, // remove extension
				AudioPath:  fmt.Sprintf("%v/%v/%v", serverAddress, dirPath, d.Name()),
				ImagePath:  fmt.Sprintf("%v/%v/%v", serverAddress, dirPath, sourceType+"_spectrogram.png"),
			})
		}
		return nil
	})

	return res
}

func generateSpectrograms(dirName string) error {
	dirPath := filepath.Join(uploadsPath, dirName) // uploads/{dirName}
	errGroup := new(errgroup.Group)

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && isWav(d) {
			errGroup.Go(func() error {
				err := exec.Command(pythonPath, spectrogramScript, path).Run()
				return err
			})
		}
		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func moveDirContents(source, target string) error {
	// first create destination dir if doesn't exist
	err := os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return err
	}

	filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if err := os.Rename(path, filepath.Join(target, d.Name())); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

func main() {
	index := template.Must(template.ParseFiles("./templates/index.html"))
	afterUpload := template.Must(template.ParseFiles("./templates/after_upload.html"))
	afterSeparate := template.Must(template.ParseFiles("./templates/separation_result.html"))

	s := chi.NewRouter()

	s.Get("/uploads/{file}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /uploads/%s", chi.URLParam(r, "file"))
		http.ServeFile(w, r, fmt.Sprintf("./uploads/%s", chi.URLParam(r, "file")))
	})

	s.Get("/uploads/{folder}/{file}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /uploads/%s/%s", chi.URLParam(r, "folder"), chi.URLParam(r, "file"))
		http.ServeFile(w, r, fmt.Sprintf("./uploads/%s/%s", chi.URLParam(r, "folder"), chi.URLParam(r, "file")))
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

	s.Get("/separate-median", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /separate-median")

		startTime := time.Now()

		// treba uzet uploadani original
		// napravit separaciju -> pozvat python, skriptu separate_median.py
		// napravit spektrograme -> za svaki fajl koji nastane pozvat pajton, skriptu spectrogram.py i dat fajlname
		// a zasto ne samo generirat sve fajlove u folderu? samo mu das folder i on izgenerira spektrogram za svaki wav u njemu
		// pa da, to se cini jednostavnije, napravis neki walkdir ili neki vrag u folderu u pajton skripti i rokavela

		cmd := exec.Command(pythonPath, "./scripts/separate_median.py")
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = generateSpectrograms("median")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		d := getSeparatedData("median")

		endTime := time.Now()
		fmt.Printf("\n\nseparate-median took %v seconds\n", (endTime.Sub(startTime)).Seconds())

		afterSeparate.Execute(w, d)
	})

	s.Get("/separate-unmix", func(w http.ResponseWriter, r *http.Request) {
		log.Println("GET /separate-unmix")
		startTime := time.Now()

		cmd := exec.Command(unmixPath, "./uploads/original.wav")

		_, err := cmd.CombinedOutput() // TODO -> use run
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = moveDirContents("./original_umxl", "./uploads/umx")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = generateSpectrograms("umx")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		d := getSeparatedData("umx")

		endTime := time.Now()
		fmt.Printf("\n\nseparate-unmix took %v seconds\n", (endTime.Sub(startTime)).Seconds())

		afterSeparate.Execute(w, d)
	})

	fmt.Println("Starting server on localhost:8080")
	if err := http.ListenAndServe(":8080", s); err != nil {
		log.Fatal(err)
	}
}
