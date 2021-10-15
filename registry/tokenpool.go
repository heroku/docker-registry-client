package registry

import (
	"strings"
	"sync"
)

type TokenPool struct {
	tokens map[string]string
	rwm    *sync.RWMutex
}

func NewTokenPool() *TokenPool {
	return &TokenPool{
		tokens: make(map[string]string),
		rwm:    &sync.RWMutex{},
	}
}

func (t *TokenPool) GetToken(scope string) string {
	if scope == "" {
		return ""
	}

	if _, ok := t.tokens[scope]; ok {
		return t.tokens[scope]
	}
	return ""
}

func (t *TokenPool) SetToken(scope, token string) {
	// repository:gds-eip/eip-api:pull
	if l := strings.Split(scope, ":"); len(l) == 3 {
		t.rwm.Lock()
		t.tokens[l[1]] = token
		t.rwm.Unlock()
	}

}
