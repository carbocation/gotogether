package gotogether

import (
	"archive/zip"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	Version = "0.3.2"
)

var ResourceMap map[string]Resource = nil
var initMutex sync.Mutex

func loadMap() (map[string]Resource, error) {
	this := os.Args[0]
	file, err := os.Open(this)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	rdr, err := zip.NewReader(file, info.Size())
	if err != nil {
		return nil, err
	}

	entries := make(map[string]Resource)
	for _, file := range rdr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		entries[file.Name] = &resource{file}
	}

	return entries, nil
}

func Initialize() error {
	initMutex.Lock()
	defer initMutex.Unlock()

	if ResourceMap != nil {
		return nil
	}
	var err error
	ResourceMap, err = loadMap()
	return err
}

type Resource interface {
	Name() string
	Open() (io.ReadCloser, error)
	Size() int64
	ModTime() time.Time
}

type resource struct {
	entry *zip.File
}

func (rsc *resource) Name() string {
	return rsc.entry.Name
}

func (rsc *resource) Open() (io.ReadCloser, error) {
	return rsc.entry.Open()
}

func (rsc *resource) Size() int64 {
	return rsc.entry.FileInfo().Size()
}

func (rsc *resource) ModTime() time.Time {
	return rsc.entry.FileInfo().ModTime()
}

type handler int

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rsc := Get(req.URL.Path)
	if rsc == nil {
		http.NotFound(w, req)
		return
	}

	rdr, err := rsc.Open()
	if err != nil {
		message := fmt.Sprintf("can't open %s - %s", rsc.Name(), err)
		http.Error(w, message, http.StatusInternalServerError)
	}
	defer rdr.Close()

	mtype := mime.TypeByExtension(filepath.Ext(req.URL.Path))
	if len(mtype) != 0 {
		w.Header().Set("Content-Type", mtype)
	}
	w.Header().Set("Content-Size", fmt.Sprintf("%d", rsc.Size()))
	w.Header().Set("Last-Modified", rsc.ModTime().UTC().Format(http.TimeFormat))
	
	//Set a 1-year expiration time. These ARE static resources, right?
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", 31536000)) 

	io.Copy(w, rdr)
}

// Get returns the named resource (nil if not found)
func Get(path string) Resource {
	return ResourceMap[path]
}

// Handle register HTTP handler under prefix
func Handle(prefix string) error {
	if err := Initialize(); err != nil {
		return err
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	var h handler
	http.Handle(prefix, http.StripPrefix("/", h))
	return nil
}

// LoadTemplates loads named templates from resources.
// If the argument "t" is nil, it is created from the first resource.
func LoadTemplates(t *template.Template, filenames ...string) (*template.Template, error) {
	if err := Initialize(); err != nil {
		return nil, err
	}

	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("no files named in call to LoadTemplates")
	}

	for _, filename := range filenames {
		rsc := Get(filename)
		if rsc == nil {
			return nil, fmt.Errorf("can't find %s", filename)
		}

		rdr, err := rsc.Open()
		if err != nil {
			return nil, fmt.Errorf("can't open %s - %s", filename, err)
		}
		data, err := ioutil.ReadAll(rdr)
		if err != nil {
			return nil, err
		}

		var tmpl *template.Template
		name := filepath.Base(filename)
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(string(data))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
