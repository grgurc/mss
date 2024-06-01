package templating

import (
	"html/template"
	"io"
	"path"
)

type Templater struct {
	index         *template.Template
	afterUpload   *template.Template
	afterSeparate *template.Template
}

func NewTemplater(basePath string) *Templater {
	templatesPath := path.Join(basePath, "templates")

	indexPath := path.Join(templatesPath, "index.html")
	afterUploadPath := path.Join(templatesPath, "after_upload.html")
	afterSepPath := path.Join(templatesPath, "separation_result.html")

	return &Templater{
		index:         template.Must(template.ParseFiles(indexPath)),
		afterUpload:   template.Must(template.ParseFiles(afterUploadPath)),
		afterSeparate: template.Must(template.ParseFiles(afterSepPath)),
	}
}

// TODO: better way to do this...

type AudioData struct {
	SourceType string
	AudioPath  string
	ImagePath  string
}

type UploadData struct {
	Methods   []string
	AudioPath string
	ImagePath string
}

func (t *Templater) Index(w io.Writer) {
	t.index.Execute(w, nil)
}

func (t *Templater) AfterUpload(w io.Writer, data UploadData) {
	t.afterUpload.Execute(w, data)
}

func (t *Templater) AfterSeparate(w io.Writer, data []AudioData) {
	t.afterSeparate.Execute(w, data)
}
