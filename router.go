package gins

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	MethodGet     = "Get"
	MethodHead    = "Head"
	MethodPost    = "Post"
	MethodPut     = "Put"
	MethodPatch   = "Patch" // RFC 5789
	MethodDelete  = "Delete"
	MethodConnect = "Connect"
	MethodOptions = "Options"
	MethodTrace   = "Trace"

	StaticFileFS = "StaticFileFS"
	StaticFile   = "StaticFile"
	StaticFS     = "StaticFS"
	Static       = "Static"
)

var AnyMethods = []string{
	MethodGet,
	MethodHead,
	MethodPost,
	MethodPut,
	MethodPatch,
	MethodDelete,
	MethodConnect,
	MethodOptions,
	MethodTrace,
}

var match = append(AnyMethods, StaticFileFS, StaticFile, StaticFS, Static)
var reg = regexp.MustCompile("^(" + strings.Join(match, "|") + ")(\\w+)")

type Router struct {
	handle reflect.Value
	static map[string]reflect.Value
}

func (r *Router) Bind(name string, handle reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("method \"%s\" error: %v", name, r)
		}
	}()

	s := reg.FindStringSubmatch(name)
	if s == nil {
		return nil
	}

	if static, ok := r.static[s[1]]; ok {
		static.Call(handle.Call([]reflect.Value{}))
	} else {
		r.handle.Call([]reflect.Value{
			reflect.ValueOf(strings.ToUpper(s[1])),
			reflect.ValueOf(ParseName(s[2])),
			handle,
		})
	}
	return nil
}

func NewRouter(r *gin.RouterGroup) *Router {
	return &Router{
		handle: reflect.ValueOf(r.Handle),
		static: map[string]reflect.Value{
			StaticFileFS: reflect.ValueOf(r.StaticFileFS),
			StaticFile:   reflect.ValueOf(r.StaticFile),
			StaticFS:     reflect.ValueOf(r.StaticFS),
			Static:       reflect.ValueOf(r.Static),
		},
	}
}
