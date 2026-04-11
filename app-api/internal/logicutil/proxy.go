package logicutil

import (
	"context"
	"reflect"

	"wklive/common/utils"

	"google.golang.org/grpc"
)

func Proxy[Resp any, PReq any, PResp any](ctx context.Context, req any, call func(context.Context, *PReq, ...grpc.CallOption) (*PResp, error)) (*Resp, error) {
	protoReq := new(PReq)
	if err := copyValue(reflect.ValueOf(protoReq), reflect.ValueOf(req)); err != nil {
		return nil, err
	}

	if uid, err := utils.GetUidFromCtx(ctx); err == nil {
		setInt64Field(protoReq, uid, "UserId", "Uid")
	}

	tenantId := getInt64Field(req, "TenantId")
	if tenantId == 0 {
		if ctxTenantId, err := utils.GetTenantIdFromCtx(ctx); err == nil {
			tenantId = ctxTenantId
		}
	}
	if tenantId != 0 {
		setInt64Field(protoReq, tenantId, "TenantId")
	}

	protoResp, err := call(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	resp := new(Resp)
	if err := copyValue(reflect.ValueOf(resp), reflect.ValueOf(protoResp)); err != nil {
		return nil, err
	}

	return resp, nil
}

func copyValue(dst, src reflect.Value) error {
	if !dst.IsValid() || !src.IsValid() {
		return nil
	}

	for dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}

	for src.Kind() == reflect.Pointer {
		if src.IsNil() {
			return nil
		}
		src = src.Elem()
	}

	if !src.IsValid() || !dst.CanSet() {
		return nil
	}

	switch dst.Kind() {
	case reflect.Struct:
		if src.Kind() != reflect.Struct {
			if src.Type().AssignableTo(dst.Type()) {
				dst.Set(src)
			} else if src.Type().ConvertibleTo(dst.Type()) {
				dst.Set(src.Convert(dst.Type()))
			}
			return nil
		}
		copyStruct(dst, src)
	case reflect.Slice:
		if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
			return nil
		}
		slice := reflect.MakeSlice(dst.Type(), src.Len(), src.Len())
		for i := 0; i < src.Len(); i++ {
			if err := copyValue(slice.Index(i), src.Index(i)); err != nil {
				return err
			}
		}
		dst.Set(slice)
	default:
		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
		} else if src.Type().ConvertibleTo(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
		}
	}

	return nil
}

func copyStruct(dst, src reflect.Value) {
	srcType := src.Type()
	for i := 0; i < src.NumField(); i++ {
		sf := srcType.Field(i)
		if sf.PkgPath != "" {
			continue
		}

		srcField := src.Field(i)
		if sf.Anonymous {
			_ = copyValue(dst, srcField)
			continue
		}

		if sf.Name == "PageReq" {
			setPageField(dst, src)
			continue
		}

		targetName := sf.Name
		if targetName == "Base" {
			targetName = "RespBase"
		}

		if dstField, ok := findField(dst, targetName); ok {
			_ = copyValue(dstField, srcField)
		}
	}

	setPageField(dst, src)
}

func setPageField(dst, src reflect.Value) {
	pageField, ok := findField(dst, "Page")
	if !ok {
		return
	}

	cursorField, okCursor := findField(src, "Cursor")
	limitField, okLimit := findField(src, "Limit")
	if !okCursor || !okLimit {
		return
	}

	if pageField.Kind() == reflect.Pointer {
		if pageField.IsNil() {
			pageField.Set(reflect.New(pageField.Type().Elem()))
		}
		pageField = pageField.Elem()
	}

	if cursorDst, ok := findField(pageField, "Cursor"); ok {
		_ = copyValue(cursorDst, cursorField)
	}
	if limitDst, ok := findField(pageField, "Limit"); ok {
		_ = copyValue(limitDst, limitField)
	}
}

func setInt64Field(target any, value int64, names ...string) {
	v := reflect.ValueOf(target)
	for _, name := range names {
		if field, ok := findField(v, name); ok {
			_ = copyValue(field, reflect.ValueOf(value))
		}
	}
}

func getInt64Field(source any, name string) int64 {
	if source == nil {
		return 0
	}
	if field, ok := findField(reflect.ValueOf(source), name); ok {
		for field.Kind() == reflect.Pointer {
			if field.IsNil() {
				return 0
			}
			field = field.Elem()
		}
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(field.Uint())
		}
	}
	return 0
}

func findField(v reflect.Value, name string) (reflect.Value, bool) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.PkgPath != "" {
			continue
		}
		if fieldType.Name == name {
			return fieldValue, true
		}
		if fieldType.Anonymous {
			if nested, ok := findField(fieldValue, name); ok {
				return nested, true
			}
		}
	}

	return reflect.Value{}, false
}
