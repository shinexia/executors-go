package executors

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Cast converts the input value to the specified type T
func Cast[T any](src any) (T, error) {
	var dst T
	err := reflectCastTo(reflect.ValueOf(&src), reflect.ValueOf(&dst).Elem())
	return dst, err
}

// CastTo performs type conversion from src to dst
func CastTo[T any](src any, dst T) error {
	return reflectCastTo(reflect.ValueOf(&src).Elem(), reflect.ValueOf(dst))
}

func reflectCastTo(src reflect.Value, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Invalid:
		return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v)", dst.Kind(), src.Type(), src)
	}
	for (src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface) && !src.IsNil() {
		src = src.Elem()
	}

	for dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			if reflectIsNil(src) {
				return nil
			}
			dstType := dst.Type()
			if src.CanAddr() {
				srcAddr := src.Addr()
				if srcAddr.CanConvert(dstType) {
					dst.Set(srcAddr.Convert(dstType))
					return nil
				}
			}
			dstV := reflect.New(dstType.Elem()).Elem()
			err := reflectCastTo(src, dstV)
			if err != nil {
				return err
			}
			dst.Set(dstV.Addr())
			return nil
		}
		dst = dst.Elem()
	}
	dstType := dst.Type()
	if reflectIsNil(src) {
		dst.Set(reflect.Zero(dstType))
		return nil
	} else if src.Type() == dstType {
		dst.Set(src)
		return nil
	} else if src.CanConvert(dstType) {
		dst.Set(src.Convert(dstType))
		return nil
	}
	switch dst.Kind() {
	case reflect.Interface:
		return reflectConvertTo(src, dst)
	case reflect.Bool:
		x, err := reflectCastBool(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(bool(x)), dst)
	case reflect.Int:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(int(x)), dst)
	case reflect.Int8:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(int8(x)), dst)
	case reflect.Int16:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(int64(x)), dst)
	case reflect.Int32:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(int32(x)), dst)
	case reflect.Int64:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(int64(x)), dst)
	case reflect.Uint:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(uint(x)), dst)
	case reflect.Uint8:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(uint8(x)), dst)
	case reflect.Uint16:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(uint16(x)), dst)
	case reflect.Uint32:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(uint32(x)), dst)
	case reflect.Uint64:
		x, err := reflectCastInt64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(uint64(x)), dst)
	case reflect.Float32:
		x, err := reflectCastFloat64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(float32(x)), dst)
	case reflect.Float64:
		x, err := reflectCastFloat64(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(float64(x)), dst)
	case reflect.String:
		x, err := reflectCastString(src)
		if err != nil {
			return err
		}
		return reflectConvertTo(reflect.ValueOf(string(x)), dst)
	case reflect.Map:
		return reflectCastToMap(src, dst)
	case reflect.Array:
		return reflectCastToArray(src, dst)
	case reflect.Slice:
		return reflectCastToSlice(src, dst)
	case reflect.Struct:
		return reflectCastToStruct(src, dst)
	}
	return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v)", dst.Type(), src.Type(), src)
}

func reflectContainsField(src reflect.Value, names ...string) bool {
	for src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface {
		if src.IsNil() {
			return false
		}
		src = src.Elem()
	}
	switch src.Kind() {
	case reflect.Map:
		for _, name := range names {
			x := src.MapIndex(reflect.ValueOf(name))
			if x.Kind() != reflect.Invalid {
				return true
			}
		}
	case reflect.Struct:
		for _, name := range names {
			x := src.FieldByName(name)
			if x.Kind() != reflect.Invalid {
				return true
			}
		}
		srcType := src.Type()
		numField := srcType.NumField()
		for i := 0; i < numField; i++ {
			field := srcType.Field(i)
			if !field.IsExported() {
				continue
			}
			tag := getJsonTag(field.Tag.Get("json"))
			for _, name := range names {
				if name == tag {
					return true
				}
			}
		}
	}
	return false
}

func reflectConvertTo(src reflect.Value, dst reflect.Value) error {
	if src.Type() == dst.Type() {
		dst.Set(src)
		return nil
	}
	if src.CanConvert(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return nil
	}
	return NewRuntimeErrorf("convert fail, dst: bool, src: %v(%v)", src.Type(), src)
}

func reflectCastBool(v reflect.Value) (bool, error) {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 1, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 1, nil
	case reflect.Float32, reflect.Float64:
		return int(math.Round(v.Float())) == 1, nil
	case reflect.String:
		return strconv.ParseBool(v.String())
	}
	return false, NewRuntimeErrorf("convert fail, dst: bool, src: %v(%v)", v.Type(), v)
}

func reflectCastInt64(v reflect.Value) (int64, error) {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(math.Round(v.Float())), nil
	case reflect.String:
		return strconv.ParseInt(v.String(), 10, 64)
	}
	return 0, NewRuntimeErrorf("convert int64 fail, v: %v(%v)", v.Type(), v)
}

func reflectCastFloat64(v reflect.Value) (float64, error) {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return 1, nil
		}
		return 0, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return strconv.ParseFloat(v.String(), 64)
	}
	return 0, NewRuntimeErrorf("convert fail, dst: float64, src: %v(%v)", v.Type(), v)
}

func reflectCastString(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Slice:
		i := v.Interface()
		d, ok := i.([]byte)
		if ok {
			return string(d), nil
		}
	}
	return "", NewRuntimeErrorf("convert fail, dst: string, src: %v(%v)", v.Type(), v)
}

func reflectCastToMap(v reflect.Value, dst reflect.Value) error {
	switch v.Kind() {
	case reflect.Map:
		iter := v.MapRange()
		dstType := dst.Type()
		keyType := dstType.Key()
		valType := dstType.Elem()
		dst.Set(reflect.MakeMapWithSize(dstType, v.Len()))
		for iter.Next() {
			dstK := reflect.New(keyType).Elem()
			err := reflectCastTo(iter.Key(), dstK)
			if err != nil {
				return err
			}
			dstV := reflect.New(valType).Elem()
			err = reflectCastTo(iter.Value(), dstV)
			if err != nil {
				return err
			}
			dst.SetMapIndex(dstK, dstV)
		}
		return nil
	case reflect.Struct:
		srcType := v.Type()
		dstType := dst.Type()
		dstKeyType := dstType.Key()
		dstValType := dst.Type().Elem()
		numField := srcType.NumField()
		for i := 0; i < numField; i++ {
			field := srcType.Field(i)
			if !field.IsExported() {
				continue
			}
			tag := getJsonTag(field.Tag.Get("json"))
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = field.Name
			}
			dstK := reflect.New(dstKeyType).Elem()
			err := reflectCastTo(dstK, reflect.ValueOf(tag))
			if err != nil {
				return err
			}
			dstV := reflect.New(dstValType).Elem()
			err = reflectCastTo(dstV, v.Field(i))
			if err != nil {
				return err
			}
			dst.SetMapIndex(dstK, dstV)
		}
	}
	return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v)", dst.Type(), v.Kind(), v)
}

func reflectCastToArray(v reflect.Value, dst reflect.Value) error {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		dstLen := dst.Len()
		srcLen := v.Len()
		if srcLen == 0 {
			valType := dst.Type().Elem()
			zero := reflect.Zero(valType)
			for i := 0; i < dstLen; i++ {
				dst.Index(i).Set(zero)
			}
			return nil
		}
		if srcLen != dstLen {
			return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v), invalid length", dst.Type(), v.Kind(), v)
		}
		for i := 0; i < srcLen; i++ {
			err := reflectCastTo(v.Index(i), dst.Index(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
	return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v)", dst.Type(), v.Kind(), v)
}

func reflectCastToSlice(v reflect.Value, dst reflect.Value) error {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		dstType := dst.Type()
		srcLen := v.Len()
		if dst.Cap() < srcLen {
			dst.Set(reflect.MakeSlice(dstType, srcLen, srcLen))
		} else {
			dstLen := dst.Len()
			if dstLen > srcLen {
				elmType := dstType.Elem()
				zero := reflect.Zero(elmType)
				for i := srcLen; i < dstLen; i++ {
					dst.Index(i).Set(zero)
				}
			}
			dst.SetLen(srcLen)
		}
		if srcLen == 0 {
			return nil
		}
		for i := 0; i < srcLen; i++ {
			err := reflectCastTo(v.Index(i), dst.Index(i))
			if err != nil {
				return err
			}
		}
		return nil
	}
	return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v)", dst.Type(), v.Kind(), v)
}

func reflectCastToStruct(v reflect.Value, dst reflect.Value) error {
	switch v.Kind() {
	case reflect.Map:
		vKeyType := v.Type().Key()
		if !reflect.TypeOf((*string)(nil)).Elem().ConvertibleTo(vKeyType) {
			return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v), type of src's key not string", dst.Type(), v.Kind(), v)
		}
		dstType := dst.Type()
		numField := dstType.NumField()
		for i := 0; i < numField; i++ {
			field := dstType.Field(i)
			if !field.IsExported() {
				continue
			}
			tag := getJsonTag(field.Tag.Get("json"))
			if tag == "-" {
				continue
			}
			if tag == "" {
				tag = field.Name
			}
			srcV := v.MapIndex(reflect.ValueOf(tag).Convert(vKeyType))
			err := reflectCastTo(srcV, dst.Field(i))
			if err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		var addr = dst.Addr().Interface()
		str := v.String()
		if dst.Type() == reflect.TypeOf((*time.Time)(nil)).Elem() {
			str = "\"" + str + "\""
		}
		err := json.Unmarshal([]byte(str), addr)
		if err != nil {
			return NewRuntimeErrorf("convert fail, dst: %v, src: string(%v), err: %v", dst.Type(), str, err)
		}
		dst.Set(reflect.ValueOf(addr).Elem())
		return nil
	}
	return NewRuntimeErrorf("convert fail, dst: %v, src: %v(%v)", dst.Type(), v.Type(), v)
}

func reflectIsNil(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.Invalid:
		return true
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func getJsonTag(tag string) string {
	return strings.Split(tag, ",")[0]
}
