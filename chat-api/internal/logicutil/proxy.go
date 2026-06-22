package logicutil

import (
	"context"
	"reflect"

	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	codeForbidden          int32 = 403
	codeNotFound           int32 = 404
	codeTooManyRequests    int32 = 429
	codeServiceUnavailable int32 = 100000
	codeSystemError        int32 = 100001
	codeUnauthorized       int32 = 100002
	codeBadRequest         int32 = 100003
)

func Proxy[Resp any, PReq any, PResp any](ctx context.Context, req any, call func(context.Context, *PReq, ...grpc.CallOption) (*PResp, error)) (*Resp, error) {
	protoReq := new(PReq)
	if err := copyValue(reflect.ValueOf(protoReq), reflect.ValueOf(req)); err != nil {
		return systemErrorResp[Resp](ctx, err)
	}

	ctx = injectMetadataValues(ctx, reflect.ValueOf(req))
	protoResp, err := call(ctx, protoReq)
	if err != nil {
		resp := new(Resp)
		code, _ := rpcErrorCodeAndMessage(err)
		if code == codeSystemError {
			logx.WithContext(ctx).Errorf("proxy rpc system error: %v", err)
		}
		if setErrorResp(resp, err) {
			return resp, nil
		}
		return systemErrorResp[Resp](ctx, err)
	}

	resp := new(Resp)
	if err := copyValue(reflect.ValueOf(resp), reflect.ValueOf(protoResp)); err != nil {
		return systemErrorResp[Resp](ctx, err)
	}

	return resp, nil
}

func injectMetadataValues(ctx context.Context, req reflect.Value) context.Context {
	if userID, ok := int64Field(req, "UserId"); ok && userID > 0 {
		ctx = context.WithValue(ctx, utils.CtxKeyUid, userID)
	}
	if merchantID, ok := int64Field(req, "MerchantId"); ok && merchantID > 0 {
		ctx = context.WithValue(ctx, utils.CtxKeyMerchantId, merchantID)
	}
	if username, ok := stringField(req, "Username"); ok && username != "" {
		ctx = context.WithValue(ctx, utils.CtxKeyUsername, username)
	}
	return ctx
}

func int64Field(v reflect.Value, name string) (int64, bool) {
	field, ok := findField(v, name)
	if !ok {
		return 0, false
	}
	for field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return 0, false
		}
		field = field.Elem()
	}
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(field.Uint()), true
	default:
		return 0, false
	}
}

func stringField(v reflect.Value, name string) (string, bool) {
	field, ok := findField(v, name)
	if !ok {
		return "", false
	}
	for field.Kind() == reflect.Pointer {
		if field.IsNil() {
			return "", false
		}
		field = field.Elem()
	}
	if field.Kind() != reflect.String {
		return "", false
	}
	return field.String(), true
}

func systemErrorResp[Resp any](ctx context.Context, err error) (*Resp, error) {
	logx.WithContext(ctx).Errorf("logic system error: %v", err)

	resp := new(Resp)
	_ = setRespCodeAndMsg(reflect.ValueOf(resp), codeSystemError, err.Error())
	return resp, nil
}

func setErrorResp(resp any, err error) bool {
	code, msg := rpcErrorCodeAndMessage(err)
	v := reflect.ValueOf(resp)
	if setRespCodeAndMsg(v, code, msg) {
		return true
	}

	base, ok := findField(v, "RespBase")
	if !ok {
		base, ok = findField(v, "Base")
	}
	if !ok {
		return false
	}
	return setRespCodeAndMsg(base, code, msg)
}

func setRespCodeAndMsg(v reflect.Value, code int32, msg string) bool {
	codeField, okCode := findField(v, "Code")
	msgField, okMsg := findField(v, "Msg")
	if !okCode || !okMsg || !codeField.CanSet() || !msgField.CanSet() {
		return false
	}
	codeField.SetInt(int64(code))
	msgField.SetString(msg)
	return true
}

func rpcErrorCodeAndMessage(err error) (int32, string) {
	st, ok := status.FromError(err)
	if !ok {
		return codeSystemError, err.Error()
	}

	code := int32(st.Code())
	if code >= 1000 {
		return code, st.Message()
	}

	switch st.Code() {
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return codeBadRequest, st.Message()
	case codes.Unauthenticated:
		return codeUnauthorized, st.Message()
	case codes.PermissionDenied:
		return codeForbidden, st.Message()
	case codes.NotFound:
		return codeNotFound, st.Message()
	case codes.ResourceExhausted:
		return codeTooManyRequests, st.Message()
	case codes.DeadlineExceeded, codes.Unavailable:
		return codeServiceUnavailable, st.Message()
	default:
		return codeSystemError, st.Message()
	}
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
			setAssignable(dst, src)
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
		setAssignable(dst, src)
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

		switch sf.Name {
		case "PageReq":
			setNestedStructField(dst, src, "Page", "Cursor", "Limit")
			continue
		case "TimeRange":
			setNestedStructField(dst, src, "TimeRange", "StartTime", "EndTime")
			continue
		}

		targetName := sf.Name
		if targetName == "Base" {
			targetName = "RespBase"
		}
		if targetName == "Type" {
			targetName = "SenderType"
		}

		if dstField, ok := findField(dst, targetName); ok {
			_ = copyValue(dstField, srcField)
			continue
		}

		if sf.Name == "Base" {
			_ = copyValue(dst, srcField)
		}
	}

	setNestedStructField(dst, src, "Page", "Cursor", "Limit")
	setNestedStructField(dst, src, "TimeRange", "StartTime", "EndTime")
}

func setNestedStructField(dst, src reflect.Value, target string, fieldNames ...string) {
	targetField, ok := findField(dst, target)
	if !ok {
		return
	}

	for targetField.Kind() == reflect.Pointer {
		if targetField.IsNil() {
			targetField.Set(reflect.New(targetField.Type().Elem()))
		}
		targetField = targetField.Elem()
	}
	if targetField.Kind() != reflect.Struct {
		return
	}

	for _, name := range fieldNames {
		srcField, okSrc := findField(src, name)
		dstField, okDst := findField(targetField, name)
		if okSrc && okDst {
			_ = copyValue(dstField, srcField)
		}
	}
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

func setAssignable(dst, src reflect.Value) {
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
	} else if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
	}
}
