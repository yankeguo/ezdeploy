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

type Database struct {
	blob *ezblob.Blob
	lock *sync.RWMutex
	data map[string]string
}

// Open create an instance and load existed data / 创建一个实例并载入已有的数据
func Open(ctx context.Context, opts Options) (db *Database, err error) {
	var blob *ezblob.Blob
	if blob, err = ezblob.New(ezblob.Options(opts)); err != nil {
		return
	}
	db = &Database{
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
		var gr *gzip.Reader
		if gr, err = gzip.NewReader(bytes.NewReader(buf)); err != nil {
			return
		}
		if err = gob.NewDecoder(gr).Decode(&db.data); err != nil {
			return
		}
	}
	return
}

// Put set a key-value / 设置一个键值对
func (db *Database) Put(key string, val string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.data[key] = val
	return
}

// Get retrieve value by key / 读取一个值
func (db *Database) Get(key string) string {
	db.lock.RLock()
	defer db.lock.RUnlock()
	return db.data[key]
}

// Del delete a value by key / 移除一个值
func (db *Database) Del(key string) {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.data, key)
}

// Purge iterate all entries and determine whether to delete / 遍历所有键值对，并决定是否移除某个键值
func (db *Database) Purge(fn func(key string, val string) (del bool, stop bool)) {
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
	return
}

// Serialize marshal internal data / 序列化内部数据
func (db *Database) Serialize() (data []byte, err error) {
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

// Save persist data to kubernetes / 将内部数据保存到 Kubernetes
func (db *Database) Save(ctx context.Context) (err error) {
	var buf []byte
	if buf, err = db.Serialize(); err != nil {
		return
	}
	if err = db.blob.Save(ctx, buf); err != nil {
		return
	}
	return
}
