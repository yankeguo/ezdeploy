package ezblob

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"strconv"
	"sync"
)

const (
	LabelKeyManagedBy = "app.kubernetes.io/managed-by"
	LabelValManagedBy = "ezblob.guoyk93.github.io"
	LabelKeyComponent = "app.kubernetes.io/component"
	LabelValComponent = "chunk"
	LabelKeyName      = "guoyk93.github.io/ezblob-name"
	LabelKeyRevision  = "guoyk93.github.io/ezblob-revision"

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

type headerType struct {
	Name     string
	Chunks   int
	Checksum string
	Revision string
}

func (header headerType) ToData() map[string][]byte {
	return map[string][]byte{
		KeyName:     []byte(header.Name),
		KeyRevision: []byte(header.Revision),
		KeyChecksum: []byte(header.Checksum),
		KeyChunks:   []byte(strconv.Itoa(header.Chunks)),
	}
}

// Options options to create Blob / 创建 Blob 所需要的参数
type Options struct {
	// Client kubernetes client / kubernetes 客户端
	Client *kubernetes.Clientset
	// Name Blob name, use as Secret name / Blob 名称，同时会用作 Secret 名
	Name string
	// Namespace kubernetes namespace / kubernetes 命名空间
	Namespace string
	// ChunkSize maximum size of each chunk / 最大分片大小
	ChunkSize int
}

type Blob struct {
	opts Options
	lock sync.Locker
}

// New create a Blob / 创建一个 Blob
func New(opts Options) *Blob {
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = DefaultChunkSize
	}
	if opts.Client == nil {
		panic(errors.New("ezblob: missing argument Options.Client"))
	}
	if opts.Name == "" {
		panic(errors.New("ezblob: missing argument Options.Name"))
	}
	return &Blob{opts: opts, lock: &sync.Mutex{}}
}

func (bl *Blob) apiSecret() clientcorev1.SecretInterface {
	return bl.opts.Client.CoreV1().Secrets(bl.opts.Namespace)
}

func (bl *Blob) headerGet(ctx context.Context) (header headerType, err error) {
	var secret *corev1.Secret
	if secret, err = bl.apiSecret().Get(ctx, bl.opts.Name, metav1.GetOptions{}); err != nil {
		return
	}
	header.Name = string(secret.Data[KeyName])
	if header.Name != bl.opts.Name {
		err = ErrInvalidHeaderFieldName
		return
	}
	header.Revision = string(secret.Data[KeyRevision])
	if header.Revision == "" {
		err = ErrInvalidHeaderFieldRevision
		return
	}
	if header.Chunks, err = strconv.Atoi(string(secret.Data[KeyChunks])); err != nil {
		err = ErrInvalidHeaderFieldChunks
		return
	}
	if header.Chunks < 0 {
		err = ErrInvalidHeaderFieldChunks
		return
	}
	header.Checksum = string(secret.Data[KeyChecksum])
	if header.Checksum == "" {
		err = ErrInvalidHeaderFieldChecksum
		return
	}
	return
}

func (bl *Blob) headerCreate(ctx context.Context, header headerType) (err error) {
	if _, err = bl.apiSecret().Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: bl.opts.Name,
			},
			Data: header.ToData(),
			Type: corev1.SecretTypeOpaque,
		},
		metav1.CreateOptions{},
	); err != nil {
		return
	}
	return
}

func (bl *Blob) headerPatch(ctx context.Context, header headerType) (err error) {
	var data []byte
	if data, err = json.Marshal(corev1.Secret{Data: header.ToData()}); err != nil {
		return
	}
	if _, err = bl.apiSecret().Patch(
		ctx,
		bl.opts.Name,
		types.StrategicMergePatchType,
		data,
		metav1.PatchOptions{},
	); err != nil {
		return
	}
	return
}

func (bl *Blob) headerDelete(ctx context.Context) error {
	return bl.apiSecret().Delete(ctx, bl.opts.Name, metav1.DeleteOptions{})
}

func (bl *Blob) chunkSelector() string {
	return fmt.Sprintf(
		"%s = %s, %s = %s, %s = %s",
		LabelKeyManagedBy, LabelValManagedBy,
		LabelKeyComponent, LabelValComponent,
		LabelKeyName, bl.opts.Name,
	)
}

func (bl *Blob) chunkSelectorRevision(revision string) string {
	return fmt.Sprintf(
		"%s = %s, %s = %s, %s = %s, %s = %s",
		LabelKeyManagedBy, LabelValManagedBy,
		LabelKeyComponent, LabelValComponent,
		LabelKeyName, bl.opts.Name,
		LabelKeyRevision, revision,
	)
}

func (bl *Blob) chunkSelectorNotRevision(revision string) string {
	return fmt.Sprintf(
		"%s = %s, %s = %s, %s = %s, %s != %s",
		LabelKeyManagedBy, LabelValManagedBy,
		LabelKeyComponent, LabelValComponent,
		LabelKeyName, bl.opts.Name,
		LabelKeyRevision, revision,
	)
}

func (bl *Blob) chunkName(revision string, idx int) string {
	return bl.opts.Name + "-" + revision + "-" + strconv.Itoa(idx)
}

func (bl *Blob) chunkLabelsRevision(revision string) map[string]string {
	return map[string]string{
		LabelKeyManagedBy: LabelValManagedBy,
		LabelKeyComponent: LabelValComponent,
		LabelKeyName:      bl.opts.Name,
		LabelKeyRevision:  revision,
	}
}

func (bl *Blob) chunkGet(ctx context.Context, revision string, index int) (buf []byte, err error) {
	var secret *corev1.Secret
	if secret, err = bl.apiSecret().Get(ctx, bl.chunkName(revision, index), metav1.GetOptions{}); err != nil {
		return
	}
	buf = secret.Data[KeyData]
	return
}

func (bl *Blob) chunkCreate(ctx context.Context, revision string, index int, data []byte) (err error) {
	if _, err = bl.apiSecret().Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   bl.chunkName(revision, index),
			Labels: bl.chunkLabelsRevision(revision),
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

func (bl *Blob) chunkDeleteBySelector(ctx context.Context, sel string) error {
	return bl.apiSecret().DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: sel})
}

func (bl *Blob) Delete(ctx context.Context) (err error) {
	if err = bl.apiSecret().Delete(ctx, bl.opts.Name, metav1.DeleteOptions{}); err != nil {
		return
	}
	if err = bl.chunkDeleteBySelector(ctx, bl.chunkSelector()); err != nil {
		return
	}
	return
}

// Load load all data from kubernetes secrets / 从 Kubernetes Secret 载入全部数据
func (bl *Blob) Load(ctx context.Context) (buf []byte, err error) {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	var header headerType
	if header, err = bl.headerGet(ctx); err != nil {
		if k8serrors.IsNotFound(err) {
			err = ErrNotFound
		}
		return
	}
	h := md5.New()
	for i := 0; i < header.Chunks; i++ {
		var chunk []byte
		if chunk, err = bl.chunkGet(ctx, header.Revision, i); err != nil {
			return
		}
		h.Write(chunk)
		buf = append(buf, chunk...)
	}
	if header.Checksum != hex.EncodeToString(h.Sum(nil)) {
		err = ErrChecksumMismatch
		return
	}
	return
}

// Save save data to kubernetes secrets / 把数据保存到 Kubernetes Secret
func (bl *Blob) Save(ctx context.Context, buf []byte) (err error) {
	bl.lock.Lock()
	defer bl.lock.Unlock()

	// get or create header secret
	var header headerType
	if header, err = bl.headerGet(ctx); err != nil {
		if k8serrors.IsNotFound(err) {
			// reset header
			header = headerType{
				Name: bl.opts.Name,
			}
			// create secret
			if err = bl.headerCreate(ctx, header); err != nil {
				return
			}
			// in case of error, delete created header secret
			defer func() {
				if err == nil {
					return
				}
				_ = bl.headerDelete(ctx)
			}()
		} else {
			return
		}
	}

	// create new revision
	oldRevision := header.Revision
	for {
		if header.Revision = randomRevision(); header.Revision != oldRevision {
			break
		}
	}

	chunks := splitBytes(buf, bl.opts.ChunkSize)

	// calculate chunks
	header.Chunks = len(chunks)

	// in case of error, delete created chunk secret
	defer func() {
		if err == nil {
			return
		}
		_ = bl.chunkDeleteBySelector(ctx, bl.chunkSelectorRevision(header.Revision))
	}()

	// delete chunk secret
	if err = bl.chunkDeleteBySelector(ctx, bl.chunkSelectorRevision(header.Revision)); err != nil {
		return
	}

	// create chunks
	for i, data := range chunks {
		if err = bl.chunkCreate(ctx, header.Revision, i, data); err != nil {
			return
		}
	}

	// calculate checksum
	h := md5.New()
	h.Write(buf)
	header.Checksum = hex.EncodeToString(h.Sum(nil))

	// patch header service
	if err = bl.headerPatch(ctx, header); err != nil {
		return
	}

	// delete old revision
	_ = bl.chunkDeleteBySelector(ctx, bl.chunkSelectorNotRevision(header.Revision))
	return
}
