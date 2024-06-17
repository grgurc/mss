package python

import (
	"log"
	"os/exec"
	"path"
)

type Python struct {
	execPath        string // path to python executable
	scriptsPath     string // path to scripts folder
	uploadsPath     string // path to uploads folder
	spectrogramPath string // path to spectrogram.py
}

func NewPython(basePath string) *Python {
	execPath := path.Join(basePath, "env", "bin", "python3")
	scriptsPath := path.Join(basePath, "scripts")
	uploadsPath := path.Join(basePath, "uploads")
	spectrogramPath := path.Join(scriptsPath, "spectrogram.py")

	return &Python{
		execPath:        execPath,
		scriptsPath:     scriptsPath,
		uploadsPath:     uploadsPath,
		spectrogramPath: spectrogramPath,
	}
}

func (p *Python) Spectrogram(wavPath string) error {
	cmd := exec.Command(p.execPath, p.spectrogramPath, wavPath)

	out, err := cmd.CombinedOutput()
	log.Println(string(out))
	log.Println(err)
	if err != nil {
		return err
	}

	return nil
}

func (p *Python) Separate(method string) error {
	script := path.Join(p.scriptsPath, "separate", method+".py")

	cmd := exec.Command(p.execPath, script, p.uploadsPath)

	out, err := cmd.CombinedOutput()
	log.Println(string(out))
	log.Println(err)
	if err != nil {
		return err
	}

	return nil
}
