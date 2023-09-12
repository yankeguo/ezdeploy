package ezlog

import (
	"bytes"
	"io"
	"log"
	"strings"
	"sync"
)

type LogWriter struct {
	logger *log.Logger
	buf    *bytes.Buffer
	prefix string
	lock   sync.Locker
}

func NewLogWriter(logger *log.Logger, prefix string) io.WriteCloser {
	return &LogWriter{
		logger: logger,
		prefix: prefix,
		buf:    &bytes.Buffer{},
		lock:   &sync.Mutex{},
	}
}

func (w *LogWriter) finish(force bool) {
	var (
		err  error
		line string
	)
again:
	if line, err = w.buf.ReadString('\n'); err != nil {
		if force {
			w.logger.Println(w.prefix, strings.TrimSuffix(line, "\n"))
		} else {
			w.buf.WriteString(line)
		}
	} else {
		w.logger.Println(w.prefix, strings.TrimSuffix(line, "\n"))
		goto again
	}
}

func (w *LogWriter) Close() (err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.finish(true)

	return
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	if n, err = w.buf.Write(p); err != nil {
		return
	}

	w.finish(false)

	return
}
