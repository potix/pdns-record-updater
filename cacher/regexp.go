package cacher

import (
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"sync"
)

type regexpCacher struct {
	compiledCache map[string] *pcre.Regexp
	mutex         *sync.Mutex
}

func (r regexpCacher) getRegexp(pattern string, int flags) (regexp *pcre.Regexp, err error) {
	id = Sprintf("%x:%s", flag, string)
        r.mutex.Lock()
        defer r.mutex.Unlock()
	regexp, ok := r.compiledCache[id]
	if !ok {
		regexp, compileError := pcre.Compile(pattern, flags)
		if compileError != nil {
			return nil, fmt.Errorf("%v, pattarn = %s, flags = %x", compileError, pattern, flags)
		}
		r.compileCache[id] = &regexp
	}
	return regexp, nil
}

var globalRegexCacher *regexCacher

// GetRegexpFromCache is get regexp from cache
func GetRegexpFromCache(pattern string, flags int) (regexp *pcre.Regexp, err error) {
	return globalRegexCacher.regexpCacher(pattern, flags)
}

func init() {
	globalRegexpCacher = &RegexpManager{
		compiledCache:  make(map[string]*pcre.Regexp),
		mutex:          new(sync.Mutex),
	}
}

