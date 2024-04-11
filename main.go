package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

const pythonPath string = "./env/bin/python3"

func saveUploadedFile(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	hash := sha1.Sum(data)
	dirPath := fmt.Sprintf("./uploads/%x", hash)
	if err = os.Mkdir(dirPath, 0777); err != nil {
		if !errors.Is(err, os.ErrExist) {
			log.Println(err)
		}
	}

	filePath := fmt.Sprintf("%s/original.wav", dirPath)
	saveFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer saveFile.Close()

	if _, err := saveFile.Write(data); err != nil {
		return "", err
	}
	return filePath, nil
}

func main() {
	router := gin.Default()

	router.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			c.String(http.StatusBadRequest, "Error uploading file")
			return
		}
		filePath, err := saveUploadedFile(file)
		if err != nil {
			log.Println(err)
			c.String(http.StatusBadRequest, "Error saving file")
			return
		}

		cmd := exec.Command(pythonPath, "./scripts/spectrogram.py", filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(err)
			log.Println(string(output))
			c.String(http.StatusBadRequest, "Error generating spectrogram")
			return
		}
		log.Println(string(output))

		c.String(http.StatusOK, "File uploaded")
	})

	router.GET("/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		c.File(fmt.Sprintf("./uploads/%s/original.wav", hash))
		// here we will load a webpage consisting of the spectrogram of the original file and buttons for doing all the other functionality
		// just have to figure out how to do that with gin
	})

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "index")
	})

	router.Run(":8080")
}
