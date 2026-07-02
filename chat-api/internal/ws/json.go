package ws

import (
	"encoding/base64"
	"encoding/json"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

func marshalProtoJSON(message protoreflect.Message) ([]byte, error) {
	return json.Marshal(protoMessageJSONValue(message))
}

func protoMessageJSONValue(message protoreflect.Message) any {
	if !message.IsValid() {
		return nil
	}
	switch value := message.Interface().(type) {
	case *structpb.Struct:
		return value.AsMap()
	case *structpb.Value:
		return value.AsInterface()
	case *structpb.ListValue:
		return value.AsSlice()
	}

	out := make(map[string]any)
	message.Range(func(field protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		out[field.JSONName()] = protoFieldJSONValue(field, value)
		return true
	})
	return out
}

func protoFieldJSONValue(field protoreflect.FieldDescriptor, value protoreflect.Value) any {
	if field.IsList() {
		list := value.List()
		out := make([]any, 0, list.Len())
		for i := 0; i < list.Len(); i++ {
			out = append(out, protoScalarJSONValue(field, list.Get(i)))
		}
		return out
	}
	if field.IsMap() {
		mapValue := value.Map()
		out := make(map[string]any, mapValue.Len())
		mapValue.Range(func(key protoreflect.MapKey, item protoreflect.Value) bool {
			out[key.String()] = protoScalarJSONValue(field.MapValue(), item)
			return true
		})
		return out
	}
	return protoScalarJSONValue(field, value)
}

func protoScalarJSONValue(field protoreflect.FieldDescriptor, value protoreflect.Value) any {
	switch field.Kind() {
	case protoreflect.BoolKind:
		return value.Bool()
	case protoreflect.EnumKind:
		return int32(value.Enum())
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return int32(value.Int())
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return value.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return uint32(value.Uint())
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return value.Uint()
	case protoreflect.FloatKind:
		return float32(value.Float())
	case protoreflect.DoubleKind:
		return value.Float()
	case protoreflect.StringKind:
		return value.String()
	case protoreflect.BytesKind:
		return base64.StdEncoding.EncodeToString(value.Bytes())
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return protoMessageJSONValue(value.Message())
	default:
		return nil
	}
}
