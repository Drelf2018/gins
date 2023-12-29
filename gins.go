package gins

import (
	"reflect"

	"github.com/Drelf2018/TypeGo/Reflect"
	"github.com/gin-gonic/gin"
)

func FindUseMethod(elem reflect.Value) any {
	m, ok := elem.Type().MethodByName("Use")
	if !ok || Reflect.IsEmbeddedMethod(m) {
		return nil
	}
	return elem.Method(m.Index).Interface()
}

func UnsafeBind(r *gin.RouterGroup, elem reflect.Value) (err error) {
	// bind Use function first
	switch fn := FindUseMethod(elem).(type) {
	case func(*gin.Context):
		r.Use(fn)
	case func() []gin.HandlerFunc:
		r.Use(fn()...)
	case func(*gin.RouterGroup):
		fn(r)
	}
	// bind other methods
	router := NewRouter(r)
	for k, v := range Reflect.MethodOf(elem) {
		if Reflect.IsEmbeddedMethod(k) || k.Name == "Use" {
			continue
		}
		err = router.Bind(k.Name, v)
		if err != nil {
			return
		}
	}
	// bind fields' methods
	for i, field := range Reflect.FieldOf(elem.Type()) {
		next := r
		if path, ok := field.Tag.Lookup("router"); ok {
			if path == "-" {
				continue
			}
			next = r.Group(path)
		} else if field.Name != field.Type.Name() {
			next = r.Group(ParseName(field.Name))
		}

		err = UnsafeBind(next, elem.Field(i))
		if err != nil {
			return
		}
	}

	return nil
}

func Bind(r *gin.Engine, zero any) (*gin.Engine, error) {
	return r, UnsafeBind(&r.RouterGroup, reflect.ValueOf(zero))
}

func MustBind(r *gin.Engine, zero any) *gin.Engine {
	_ = UnsafeBind(&r.RouterGroup, reflect.ValueOf(zero))
	return r
}

func Default(zero any) (*gin.Engine, error) {
	return Bind(gin.Default(), zero)
}

func MustDefault(zero any) *gin.Engine {
	return MustBind(gin.Default(), zero)
}
