package eztmp

import (
	"os"
	"sync"
)

var (
	files     []string
	filesLock = &sync.Mutex{}
)

func recordFile(file string) {
	filesLock.Lock()
	defer filesLock.Unlock()
	files = append(files, file)
}

// ClearAll clear all tmp files
func ClearAll() {
	filesLock.Lock()
	defer filesLock.Unlock()
	for _, file := range files {
		os.Remove(file)
	}
	files = nil
}

// WriteFile write file
func WriteFile(buf []byte, suffix string) (file string, err error) {
	var f *os.File
	if f, err = os.CreateTemp("", "ezops-tmp-*"+suffix); err != nil {
		return
	}
	defer f.Close()
	file = f.Name()
	recordFile(file)
	if _, err = f.Write(buf); err != nil {
		return
	}
	return
}
