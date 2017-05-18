package cacher

import (
	"github.com/pkg/errors"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"sync"
	"fmt"
)

type regexpCacher struct {
	compiledCache map[string] *pcre.Regexp
	mutex         *sync.Mutex
}

func (r regexpCacher) getRegexp(pattern string, flags int) (regexp *pcre.Regexp, err error) {
	id := fmt.Sprintf("%x:%s", flags, pattern)
        r.mutex.Lock()
        defer r.mutex.Unlock()
	regexp, ok := r.compiledCache[id]
	if !ok {
		regexp, compileError := pcre.Compile(pattern, flags)
		if compileError != nil {
			return nil, errors.Errorf("%v, pattarn = %s, flags = %x", compileError, pattern, flags)
		}
		r.compiledCache[id] = &regexp
	}
	return regexp, nil
}

var globalRegexpCacher *regexpCacher

// GetRegexpFromCache is get regexp from cache
func GetRegexpFromCache(pattern string, flags int) (regexp *pcre.Regexp, err error) {
	return globalRegexpCacher.getRegexp(pattern, flags)
}

func init() {
	globalRegexpCacher = &regexpCacher{
		compiledCache:  make(map[string]*pcre.Regexp),
		mutex:          new(sync.Mutex),
	}
}

