package ezops

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type FileSuffixes []string

func (fs FileSuffixes) Match(path string) bool {
	for _, s := range fs {
		if strings.HasSuffix(path, s) {
			return true
		}
	}
	return false
}

func readDirNames(dir string) (names []string, err error) {
	var entries []os.FileInfo
	if entries, err = ioutil.ReadDir(dir); err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() &&
			!strings.HasPrefix(entry.Name(), ".") &&
			!strings.HasPrefix(entry.Name(), "_") {
			names = append(names, entry.Name())
		}
	}
	return
}

func streamFile(w io.Writer, filename string) (err error) {
	var f *os.File
	if f, err = os.OpenFile(filename, os.O_RDONLY, 0640); err != nil {
		return
	}
	defer f.Close()
	if _, err = io.Copy(w, f); err != nil {
		return
	}
	return
}
