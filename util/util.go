package util

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

const serverAddress string = "http://localhost:8080"

func URL(parts ...string) string {
	pJoined := strings.Join(parts, "/")
	return fmt.Sprintf("%v/%v", serverAddress, pJoined)
}

func IsWav(d fs.DirEntry) bool {
	return path.Ext(d.Name()) == ".wav"
}
