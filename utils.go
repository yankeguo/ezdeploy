package ezdeploy

import (
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/yankeguo/ezdeploy/pkg/eztmp"
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
	var entries []fs.DirEntry
	if entries, err = os.ReadDir(dir); err != nil {
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

func ConvertJSONNetFileToYAML(file string, namespace string) (outFile string, err error) {
	vm := jsonnet.MakeVM()
	vm.ExtVar("NAMESPACE", namespace)
	var raw string
	if raw, err = vm.EvaluateFile(file); err != nil {
		return
	}
	if outFile, err = eztmp.WriteFile([]byte(raw), ".yaml"); err != nil {
		return
	}
	return
}
