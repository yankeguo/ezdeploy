package ezsync

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewErrorGroup(t *testing.T) {
	eg := NewErrorGroup()
	eg.Add(errors.New("hello"))
	eg.Add(nil)
	eg.Add(errors.New("world"))
	require.Equal(t, "#0: hello; #2: world", eg.Unwrap().Error())

	eg = NewErrorGroup()
	eg.Add(nil)
	eg.Add(nil)
	require.NoError(t, eg.Unwrap())
}
