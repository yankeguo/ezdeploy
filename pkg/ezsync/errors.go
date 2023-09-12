package ezsync

import (
	"strconv"
	"strings"
	"sync"
)

type Errors []error

func (errs Errors) Error() string {
	sb := &strings.Builder{}
	for i, err := range errs {
		if err == nil {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("; ")
		}
		sb.WriteRune('#')
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString(": ")
		sb.WriteString(err.Error())
	}
	return sb.String()
}

type ErrorGroup struct {
	errs []error
	lock *sync.RWMutex
}

func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{
		lock: &sync.RWMutex{},
	}
}

func (eg *ErrorGroup) Add(err error) {
	eg.lock.Lock()
	defer eg.lock.Unlock()

	eg.errs = append(eg.errs, err)
}

func (eg *ErrorGroup) Unwrap() error {
	eg.lock.RLock()
	defer eg.lock.RUnlock()

	for _, err := range eg.errs {
		if err != nil {
			return Errors(eg.errs)
		}
	}

	return nil
}
