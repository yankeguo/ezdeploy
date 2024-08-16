package ezblob

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	LabelKeyManagedBy = "app.kubernetes.io/managed-by"
	LabelValManagedBy = "ezblob.yankeguo.github.io"
	LabelKeyComponent = "app.kubernetes.io/component"
	LabelValComponent = "chunk"
	LabelKeyName      = "yankeguo.github.io/ezblob-name"
	LabelKeyRevision  = "yankeguo.github.io/ezblob-revision"

	KeyName     = "name"
	KeyRevision = "revision"
	KeyChunks   = "chunks"
	KeyChecksum = "checksum"

	KeyData = "data"
)

const (
	DefaultChunkSize = 4096
)

var (
	ErrNotFound         = errors.New("not found")
	ErrChecksumMismatch = errors.New("checksum mismatch")

	ErrInvalidHeaderFieldName     = errors.New("missing or invalid field in header secret: 'name'")
	ErrInvalidHeaderFieldChecksum = errors.New("missing or invalid field in header secret: 'checksum'")
	ErrInvalidHeaderFieldRevision = errors.New("missing or invalid field in header secret: 'revision'")
	ErrInvalidHeaderFieldChunks   = errors.New("missing or invalid field in header secret: 'chunks'")
)

type blobHeader struct {
	Name     string
	Chunks   int
	Checksum string
	Revision string
}

func (h blobHeader) ToData() map[string][]byte {
	return map[string][]byte{
		KeyName:     []byte(h.Name),
		KeyRevision: []byte(h.Revision),
		KeyChecksum: []byte(h.Checksum),
		KeyChunks:   []byte(strconv.Itoa(h.Chunks)),
	}
}

// Options Blob options
type Options struct {
	// Client kubernetes client
	Client *kubernetes.Clientset
	// Name Blob name, also used as Secret name
	Name string
	// Namespace kubernetes namespace
	Namespace string
	// ChunkSize maximum size of each chunk
	ChunkSize int
}

type Blob struct {
	client    *kubernetes.Clientset
	name      string
	namespace string
	chunkSize int
	lock      sync.Locker
}

// New create a Blob
func New(opts Options) (blob *Blob, err error) {
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = DefaultChunkSize
	}
	if opts.Client == nil {
		err = errors.New("ezblob: missing argument Options.Client")
		return
	}
	if opts.Name == "" {
		err = errors.New("ezblob: missing argument Options.Name")
		return
	}
	blob = &Blob{
		client:    opts.Client,
		name:      opts.Name,
		namespace: opts.Namespace,
		chunkSize: opts.ChunkSize,
		lock:      &sync.Mutex{},
	}
	return
}

func (b *Blob) apiSecret() clientcorev1.SecretInterface {
	return b.client.CoreV1().Secrets(b.namespace)
}

func (b *Blob) headerGet(ctx context.Context) (h blobHeader, err error) {
	var secret *corev1.Secret
	if secret, err = b.apiSecret().Get(ctx, b.name, metav1.GetOptions{}); err != nil {
		return
	}
	h.Name = string(secret.Data[KeyName])
	if h.Name != b.name {
		err = ErrInvalidHeaderFieldName
		return
	}
	h.Revision = string(secret.Data[KeyRevision])
	if h.Revision == "" {
		err = ErrInvalidHeaderFieldRevision
		return
	}
	if h.Chunks, err = strconv.Atoi(string(secret.Data[KeyChunks])); err != nil {
		err = ErrInvalidHeaderFieldChunks
		return
	}
	if h.Chunks < 0 {
		err = ErrInvalidHeaderFieldChunks
		return
	}
	h.Checksum = string(secret.Data[KeyChecksum])
	if h.Checksum == "" {
		err = ErrInvalidHeaderFieldChecksum
		return
	}
	return
}

func (b *Blob) headerCreate(ctx context.Context, h blobHeader) (err error) {
	if _, err = b.apiSecret().Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: b.name,
			},
			Data: h.ToData(),
			Type: corev1.SecretTypeOpaque,
		},
		metav1.CreateOptions{},
	); err != nil {
		return
	}
	return
}

func (b *Blob) headerPatch(ctx context.Context, h blobHeader) (err error) {
	var data []byte
	if data, err = json.Marshal(corev1.Secret{Data: h.ToData()}); err != nil {
		return
	}
	if _, err = b.apiSecret().Patch(
		ctx,
		b.name,
		types.StrategicMergePatchType,
		data,
		metav1.PatchOptions{},
	); err != nil {
		return
	}
	return
}

func (b *Blob) headerDelete(ctx context.Context) error {
	return b.apiSecret().Delete(ctx, b.name, metav1.DeleteOptions{})
}

func (b *Blob) chunkSelector() string {
	return fmt.Sprintf(
		"%s = %s, %s = %s, %s = %s",
		LabelKeyManagedBy, LabelValManagedBy,
		LabelKeyComponent, LabelValComponent,
		LabelKeyName, b.name,
	)
}

func (b *Blob) chunkSelectorRevision(revision string) string {
	return fmt.Sprintf(
		"%s = %s, %s = %s, %s = %s, %s = %s",
		LabelKeyManagedBy, LabelValManagedBy,
		LabelKeyComponent, LabelValComponent,
		LabelKeyName, b.name,
		LabelKeyRevision, revision,
	)
}

func (b *Blob) chunkSelectorNotRevision(revision string) string {
	return fmt.Sprintf(
		"%s = %s, %s = %s, %s = %s, %s != %s",
		LabelKeyManagedBy, LabelValManagedBy,
		LabelKeyComponent, LabelValComponent,
		LabelKeyName, b.name,
		LabelKeyRevision, revision,
	)
}

func (b *Blob) chunkName(revision string, idx int) string {
	return b.name + "-" + revision + "-" + strconv.Itoa(idx)
}

func (b *Blob) chunkLabelsRevision(revision string) map[string]string {
	return map[string]string{
		LabelKeyManagedBy: LabelValManagedBy,
		LabelKeyComponent: LabelValComponent,
		LabelKeyName:      b.name,
		LabelKeyRevision:  revision,
	}
}

func (b *Blob) chunkGet(ctx context.Context, revision string, index int) (buf []byte, err error) {
	var secret *corev1.Secret
	if secret, err = b.apiSecret().Get(ctx, b.chunkName(revision, index), metav1.GetOptions{}); err != nil {
		return
	}
	buf = secret.Data[KeyData]
	return
}

func (b *Blob) chunkCreate(ctx context.Context, revision string, index int, data []byte) (err error) {
	if _, err = b.apiSecret().Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   b.chunkName(revision, index),
			Labels: b.chunkLabelsRevision(revision),
		},
		Data: map[string][]byte{
			KeyData: data,
		},
		Type: corev1.SecretTypeOpaque,
	}, metav1.CreateOptions{}); err != nil {
		return
	}
	return
}

func (b *Blob) chunkDeleteBySelector(ctx context.Context, sel string) error {
	return b.apiSecret().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: sel})
}

func (b *Blob) Delete(ctx context.Context) (err error) {
	if err = b.apiSecret().Delete(ctx, b.name, metav1.DeleteOptions{}); err != nil {
		return
	}
	if err = b.chunkDeleteBySelector(ctx, b.chunkSelector()); err != nil {
		return
	}
	return
}

// Load load all data from kubernetes secrets
func (b *Blob) Load(ctx context.Context) (buf []byte, err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	var h blobHeader
	if h, err = b.headerGet(ctx); err != nil {
		if k8s_errors.IsNotFound(err) {
			err = ErrNotFound
		}
		return
	}
	d := md5.New()
	for i := 0; i < h.Chunks; i++ {
		var chunk []byte
		if chunk, err = b.chunkGet(ctx, h.Revision, i); err != nil {
			return
		}
		d.Write(chunk)
		buf = append(buf, chunk...)
	}
	if h.Checksum != hex.EncodeToString(d.Sum(nil)) {
		err = ErrChecksumMismatch
		return
	}
	return
}

// Save save data to kubernetes secrets
func (b *Blob) Save(ctx context.Context, buf []byte) (err error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	// get or create header secret
	var h blobHeader
	if h, err = b.headerGet(ctx); err != nil {
		if k8s_errors.IsNotFound(err) {
			// reset header
			h = blobHeader{
				Name: b.name,
			}
			// create secret
			if err = b.headerCreate(ctx, h); err != nil {
				return
			}
			// in case of error, delete created header secret
			defer func() {
				if err == nil {
					return
				}
				_ = b.headerDelete(ctx)
			}()
		} else {
			return
		}
	}

	// create new revision
	oldRevision := h.Revision
	for {
		if h.Revision, err = randomRevision(); err != nil {
			return
		}
		if h.Revision != oldRevision {
			break
		}
	}

	chunks := chunkify(buf, b.chunkSize)

	// calculate chunks
	h.Chunks = len(chunks)

	// in case of error, delete created chunk secret
	defer func() {
		if err == nil {
			return
		}
		_ = b.chunkDeleteBySelector(ctx, b.chunkSelectorRevision(h.Revision))
	}()

	// delete chunk secret
	if err = b.chunkDeleteBySelector(ctx, b.chunkSelectorRevision(h.Revision)); err != nil {
		return
	}

	// create chunks
	for i, data := range chunks {
		if err = b.chunkCreate(ctx, h.Revision, i, data); err != nil {
			return
		}
	}

	// calculate checksum
	d := md5.New()
	d.Write(buf)
	h.Checksum = hex.EncodeToString(d.Sum(nil))

	// patch header service
	if err = b.headerPatch(ctx, h); err != nil {
		return
	}

	// delete old revision
	_ = b.chunkDeleteBySelector(ctx, b.chunkSelectorNotRevision(h.Revision))
	return
}
