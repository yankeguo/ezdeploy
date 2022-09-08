package ezops

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/karrick/godirwalk"
	"sort"
	"strings"
)

func checksumBytes(buf []byte) string {
	h := md5.New()
	h.Write(buf)
	return hex.EncodeToString(h.Sum(nil))
}

func checksumFile(filename string) (checksum string, err error) {
	h := md5.New()
	if err = streamFile(h, filename); err != nil {
		return
	}
	checksum = hex.EncodeToString(h.Sum(nil))
	return
}

func checksumDir(dir string) (checksum string, err error) {
	var filenames []string

	if err = godirwalk.Walk(dir, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Callback: func(filename string, entry *godirwalk.Dirent) error {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
				return godirwalk.SkipThis
			}
			if entry.IsRegular() && !strings.HasPrefix(entry.Name(), ".") {
				filenames = append(filenames, filename)
			}
			return nil
		},
	}); err != nil {
		return
	}

	sort.Strings(filenames)

	h := md5.New()
	for _, filename := range filenames {
		if err = streamFile(h, filename); err != nil {
			return
		}
		h.Write([]byte{'\r', '\n'})
	}

	checksum = hex.EncodeToString(h.Sum(nil))

	return
}
