package vial

import (
	"container/list"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func getConcreteType(data reflect.Type) (reflect.Type, int) {
	level := 0
	for data.Kind() == reflect.Pointer {
		data = data.Elem()
		level++
	}
	return data, level
}

func getQualifiedClassName(data reflect.Type) string {
	data, level := getConcreteType(data)
	var builder strings.Builder
	for i := 0; i < level; i++ {
		builder.WriteByte('*')
	}
	builder.WriteString(data.PkgPath() + "." + data.Name())
	return builder.String()
}

func printCycleInjectionLoop(name string, l *list.List) string {
	e := l.Front()
	if e == nil {
		return ""
	}
	var msg strings.Builder
	msg.WriteString("Cycle Injection Found in Vial:\n")
	for e != nil {
		msg.WriteString(fmt.Sprintf("[%v] rely on [%v]\n", name, e.Value.(string)))
		name = e.Value.(string)
		e = e.Next()
	}
	return msg.String()
}

func validateDefaultValue(fieldType reflect.Type, value string) (val reflect.Value, errL error) {
	if fieldType.PkgPath() != "" {
		return reflect.Value{}, fmt.Errorf("value inject can only support original data type, won't support %v",
			getQualifiedClassName(fieldType))
	}
	concreteType, _ := getConcreteType(fieldType)
	kind := concreteType.Kind()
	defer func() {
		if errL == nil {
			val = newValue(fieldType, val)
		}
	}()
	switch kind {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int:
		res, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int(res)), nil
	case reflect.Int8:
		res, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int8(res)), nil
	case reflect.Int16:
		res, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int16(res)), nil
	case reflect.Int32:
		res, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int32(res)), nil
	case reflect.Int64:
		res, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(res), nil
	case reflect.Uint:
		res, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint(res)), nil
	case reflect.Uint8:
		res, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint8(res)), nil
	case reflect.Uint16:
		res, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint16(res)), nil
	case reflect.Uint32:
		res, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint32(res)), nil
	case reflect.Uint64:
		res, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(res), nil
	case reflect.Bool:
		res, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(res), nil
	case reflect.Float32:
		res, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(float32(res)), nil
	case reflect.Float64:
		res, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(res), nil
	case reflect.Complex64:
		res, err := strconv.ParseComplex(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(complex64(res)), nil
	case reflect.Complex128:
		res, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(res), nil
	default:
		return reflect.Value{}, fmt.Errorf("data type %v cannot use value tag", kind)
	}
}

func getKindType(dataType reflect.Type) kindType {
	dataType, _ = getConcreteType(dataType)
	switch dataType.Kind() {
	case reflect.Interface:
		return interfaceKind
	default:
		return structKind
	}
}

func newValue(targetType reflect.Type, injectValue reflect.Value) reflect.Value {
	if targetType.Kind() == reflect.Pointer {
		valPtr := reflect.New(targetType.Elem())
		elem := newValue(targetType.Elem(), injectValue)
		valPtr.Elem().Set(elem)
		return valPtr
	}
	return injectValue
}

func newValueByInject(targetType reflect.Type, injectValues []reflect.Value) reflect.Value {
	if targetType.Kind() == reflect.Pointer {
		valPtr := reflect.New(targetType.Elem())
		elem := newValueByInject(targetType.Elem(), injectValues)
		valPtr.Elem().Set(elem)
		return valPtr
	}
	ptr := 0
	elem := reflect.New(targetType).Elem()
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		_, exist1 := field.Tag.Lookup(autoWire)
		_, exist2 := field.Tag.Lookup(value)
		if exist1 || exist2 {
			elem.Field(i).Set(injectValues[ptr])
			ptr++
		}
	}
	return elem
}
