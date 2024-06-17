package separator

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/grgurcz/mss/python"
	"github.com/grgurcz/mss/util"
	"golang.org/x/sync/errgroup"
)

type Separator struct {
	python        *python.Python
	serverAddress string
	uploadsPath   string
}

func NewSeparator(basePath, serverAddress string, python *python.Python) *Separator {
	uploadsPath := path.Join(basePath, "uploads")

	return &Separator{
		uploadsPath:   uploadsPath,
		serverAddress: serverAddress,
		python:        python,
	}
}

func (s *Separator) Separate(method string) error {
	err := s.createUploadFolder(method)
	if err != nil {
		return err
	}

	if err = s.python.Separate(method); err != nil {
		return err
	}

	err = s.generateSpectrograms(method)
	if err != nil {
		return err
	}

	return nil
}

func (s *Separator) createUploadFolder(method string) error {
	p := path.Join(s.uploadsPath, method)
	err := os.MkdirAll(p, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (s *Separator) generateSpectrograms(method string) error {
	dirPath := filepath.Join(s.uploadsPath, method)
	errGroup := new(errgroup.Group)

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		fmt.Println("PATH:", path)
		if util.IsWav(d) {
			errGroup.Go(func() error {
				fmt.Println("CALLING SPECTROGRAM:", path)
				err := s.python.Spectrogram(path)
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
