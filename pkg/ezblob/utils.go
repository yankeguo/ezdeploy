package ezblob

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

func randomRevision() (s string, err error) {
	buf := make([]byte, 4)
	if _, err = rand.Read(buf); err != nil {
		return
	}
	s = strings.ToLower(
		base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf),
	)
	return
}

func chunkify(buf []byte, size int) (out [][]byte) {
	count := len(buf) / size
	if len(buf)%size != 0 {
		count = count + 1
	}
	for i := 0; i < count; i++ {
		start, end := size*i, size*(i+1)
		if end > len(buf) {
			end = len(buf)
		}
		out = append(out, buf[start:end])
	}
	return
}
