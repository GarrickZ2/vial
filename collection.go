package vial

import (
	"fmt"
	"reflect"
	"sync"
)

type singletonEntry struct {
	assigned bool
	metaInfo *structMetaInfo
	value    interface{}
	lock     sync.Mutex
}

func (s *singletonEntry) GetValue() (interface{}, error) {
	if s.assigned {
		return s.value, nil
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.assigned {
		return s.value, nil
	}
	result, err := c.buildStruct(s.metaInfo)
	if err != nil {
		return nil, err
	}
	s.assigned = true
	s.value = result
	return result, nil
}

type collection struct {
	initType     int
	singletonMap map[string]*singletonEntry
}

func newCollection() *collection {
	return &collection{0, make(map[string]*singletonEntry)}
}

// 1. if the data is an interface => find the binding
// 2. with the concrete structure, find whether
func (c *Container) getValue(data interface{}) (interface{}, error) {
	dataType := reflect.TypeOf(data)
	return c.buildStructWithSingleton(getQualifiedClassName(dataType), getKindType(dataType))
}

func (c *Container) buildStructWithSingleton(name string, kt kindType) (interface{}, error) {
	if kt == interfaceKind {
		iMetaInfo := c.register.iMap[name]
		name = iMetaInfo.primary
	}
	metaInfo := c.register.sMap[name]
	if metaInfo == nil {
		return nil, fmt.Errorf("not found %v registered in vial", name)
	}
	entry := c.collection.singletonMap[metaInfo.name]
	if entry != nil {
		return entry.GetValue()
	}
	return c.buildStruct(metaInfo)
}

func (c *Container) buildStruct(meta *structMetaInfo) (interface{}, error) {
	valueList := make([]reflect.Value, 0, len(meta.dependency))
	for _, each := range meta.dependency {
		if each.kind == valueKind {
			valueList = append(valueList, each.value)
		} else {
			buildResult, err := c.buildStructWithSingleton(each.reference, structKind)
			if err != nil {
				return nil, err
			}
			valueList = append(valueList, reflect.ValueOf(buildResult))
		}
	}
	if meta.buildType == buildByInject {
		returnResult := newValueByInject(meta.originType, valueList)
		return returnResult.Interface(), nil
	} else if meta.buildType == buildByConstructor {
		result := meta.constructor.Call(valueList)
		if len(result) == 2 {
			if result[1].Interface() != nil {
				return nil, result[1].Interface().(error)
			}
		}
		return result[0].Interface(), nil
	}
	return nil, fmt.Errorf("internal error, unknown build type")
}
