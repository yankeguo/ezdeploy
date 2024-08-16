package ezkv

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"sync"

	"github.com/yankeguo/ezdeploy/pkg/ezblob"
)

type Options ezblob.Options

type KV struct {
	blob *ezblob.Blob
	lock *sync.RWMutex
	data map[string]string
}

// Open create an instance and load existed data
func Open(ctx context.Context, opts Options) (db *KV, err error) {
	var blob *ezblob.Blob
	if blob, err = ezblob.New(ezblob.Options(opts)); err != nil {
		return
	}
	db = &KV{
		blob: blob,
		lock: &sync.RWMutex{},
		data: map[string]string{},
	}
	var buf []byte
	if buf, err = db.blob.Load(ctx); err != nil {
		if err == ezblob.ErrNotFound {
			err = nil
		} else {
			return
		}
	} else {
		if err = db.unmarshal(buf); err != nil {
			return
		}
	}
	return
}

// Put set a key-value
func (db *KV) Put(key string, val string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.data[key] = val
}

// Get retrieve value by key
func (db *KV) Get(key string) string {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.data[key]
}

// Del delete a value by key
func (db *KV) Del(key string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.data, key)
}

// Purge iterate all entries and determine whether to delete
func (db *KV) Purge(fn func(key string, val string) (del bool, stop bool)) {
	db.lock.Lock()
	defer db.lock.Unlock()
	for k, v := range db.data {
		del, stop := fn(k, v)
		if del {
			delete(db.data, k)
		}
		if stop {
			return
		}
	}
}

func (db *KV) unmarshal(buf []byte) (err error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var gr *gzip.Reader
	if gr, err = gzip.NewReader(bytes.NewReader(buf)); err != nil {
		return
	}

	if err = gob.NewDecoder(gr).Decode(&db.data); err != nil {
		return
	}
	return
}

func (db *KV) marshal() (data []byte, err error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	buf := &bytes.Buffer{}
	gw := gzip.NewWriter(buf)
	if err = gob.NewEncoder(gw).Encode(&db.data); err != nil {
		return
	}
	if err = gw.Close(); err != nil {
		return
	}
	data = buf.Bytes()
	return
}

// Save persist data to kubernetes
func (db *KV) Save(ctx context.Context) (err error) {
	var buf []byte
	if buf, err = db.marshal(); err != nil {
		return
	}
	if err = db.blob.Save(ctx, buf); err != nil {
		return
	}
	return
}
