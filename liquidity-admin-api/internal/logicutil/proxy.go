package logicutil

import (
	"context"
	"reflect"

	"google.golang.org/grpc"
)

func Proxy[Resp any, PReq any, PResp any](ctx context.Context, req any, call func(context.Context, *PReq, ...grpc.CallOption) (*PResp, error)) (*Resp, error) {
	protoReq := new(PReq)
	copyValue(reflect.ValueOf(protoReq), reflect.ValueOf(req))
	protoResp, err := call(ctx, protoReq)
	if err != nil {
		return nil, err
	}
	resp := new(Resp)
	copyValue(reflect.ValueOf(resp), reflect.ValueOf(protoResp))
	return resp, nil
}

func Convert[Resp any](src any) *Resp {
	resp := new(Resp)
	copyValue(reflect.ValueOf(resp), reflect.ValueOf(src))
	return resp
}

func copyValue(dst, src reflect.Value) {
	for dst.IsValid() && dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}
	for src.IsValid() && src.Kind() == reflect.Pointer {
		if src.IsNil() {
			return
		}
		src = src.Elem()
	}
	if !dst.IsValid() || !src.IsValid() || !dst.CanSet() {
		return
	}
	if dst.Kind() == reflect.Struct && src.Kind() == reflect.Struct {
		srcType := src.Type()
		for i := 0; i < src.NumField(); i++ {
			field := srcType.Field(i)
			if field.PkgPath != "" {
				continue
			}
			if field.Anonymous {
				copyValue(dst, src.Field(i))
				continue
			}
			target := dst.FieldByName(field.Name)
			if !target.IsValid() && field.Name == "Base" {
				target = dst.FieldByName("RespBase")
			}
			if target.IsValid() {
				copyValue(target, src.Field(i))
			}
		}
		return
	}
	if dst.Kind() == reflect.Slice && (src.Kind() == reflect.Slice || src.Kind() == reflect.Array) {
		out := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())
		for i := 0; i < src.Len(); i++ {
			copyValue(out.Index(i), src.Index(i))
		}
		dst.Set(out)
		return
	}
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
	} else if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
	}
}
