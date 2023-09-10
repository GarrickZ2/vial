package vial

import (
	"container/list"
	"fmt"
	"reflect"
)

type register struct {
	sMap map[string]*structMetaInfo
	iMap map[string]*interfaceMetaInfo
}

func newRegister() *register {
	return &register{
		sMap: make(map[string]*structMetaInfo),
		iMap: make(map[string]*interfaceMetaInfo),
	}
}

type structMetaInfo struct {
	buildType   buildType
	name        string
	option      option
	originType  reflect.Type
	constructor reflect.Value
	dependency  []*dependencyInfo
}

type dependencyInfo struct {
	name      string
	kind      kindType
	qualifier string
	reference string
	value     reflect.Value
}

type interfaceMetaInfo struct {
	primary     string
	others      map[string]bool
	nameMapping map[string]string
}

func (r *register) RegisterStruct(structure interface{}, options ...applyOption) {
	// 1. Check the first input is valid
	inputType := reflect.TypeOf(structure)
	structureType, _ := getConcreteType(inputType)
	id := getQualifiedClassName(inputType)
	if structureType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Input elem %v is not a struct related type", id))
	}

	// 2. Check the elem existence
	if _, ok := r.sMap[id]; ok {
		panic(fmt.Sprintf("Struct %v has been registered already, cannot register twice", id))
	}

	// 3. Check and register the dependency
	dependency := make([]*dependencyInfo, 0)
	for i := 0; i < structureType.NumField(); i++ {
		field := structureType.Field(i)
		if val, ok := field.Tag.Lookup(value); ok {
			if !field.IsExported() {
				panic(fmt.Sprintf("Input type %v contains field %v is unexported, cannot set as auto-wired", id, field.Name))
			}
			parseValue, parseErr := validateDefaultValue(field.Type, val)
			if parseErr != nil {
				panic("Parsing Value Tag Error: " + parseErr.Error())
			}
			name := getQualifiedClassName(field.Type)
			dependency = append(dependency, &dependencyInfo{
				name:      name,
				kind:      valueKind,
				value:     parseValue,
				reference: name,
			})
		} else if _, ok = field.Tag.Lookup(autoWire); ok {
			if !field.IsExported() {
				panic(fmt.Sprintf("Input type %v contains field %v is unexported, cannot set as auto-wired", id, field.Name))
			}
			name := getQualifiedClassName(field.Type)
			dependency = append(dependency, &dependencyInfo{
				name:      name,
				kind:      getKindType(field.Type),
				qualifier: field.Tag.Get(qualifier),
				reference: name,
			})
		}
	}

	// 4. set the options
	defaultOption := newDefaultOption()
	defaultOption.name = structureType.Name()

	for _, eachOption := range options {
		eachOption.apply(&defaultOption)
	}

	// 5. register in the map
	r.sMap[id] = &structMetaInfo{
		buildType:  buildByInject,
		name:       id,
		option:     defaultOption,
		originType: inputType,
		dependency: dependency,
	}
}

func (r *register) RegisterConstruct(constructor interface{}, options ...applyOption) {
	// 1. check input type
	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		panic(fmt.Sprintf("input elem %v is not a func", getQualifiedClassName(constructorType)))
	}
	if constructorType.NumOut() == 0 || constructorType.NumOut() > 2 {
		panic(fmt.Sprintf("constructor can only return 1 or 2 data"))
	}

	// 2.1 check return type 1
	inputType := constructorType.Out(0)
	concreteType, _ := getConcreteType(inputType)
	id := getQualifiedClassName(inputType)
	if _, ok := r.sMap[id]; ok {
		panic(fmt.Sprintf("Struct %v has been registered already, cannot register twice", id))
	}

	// 2.2 check return type 2
	if constructorType.NumOut() == 2 {
		errOut := constructorType.Out(1)
		if !errOut.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			panic("The second out type of the constructor should be error or implement error interface")
		}
	}

	// 3. Check dependency (input data)
	dependency := make([]*dependencyInfo, 0)
	for i := 0; i < constructorType.NumIn(); i++ {
		inputField := constructorType.In(i)
		name := getQualifiedClassName(inputField)
		dependency = append(dependency, &dependencyInfo{
			name:      name,
			kind:      getKindType(inputField),
			reference: name,
		})
	}

	// 4. apply the option
	defaultOption := newDefaultOption()
	defaultOption.name = concreteType.Name()
	for _, eachOption := range options {
		eachOption.apply(&defaultOption)
	}

	// 5. add to the map
	r.sMap[id] = &structMetaInfo{
		buildType:   buildByConstructor,
		name:        id,
		option:      defaultOption,
		originType:  inputType,
		constructor: reflect.ValueOf(constructor),
		dependency:  dependency,
	}
}

func (r *register) Bind(i interface{}, primaryStruct interface{}, others ...interface{}) {
	interfaceType := reflect.TypeOf(i)
	if interfaceType.Kind() == reflect.Pointer {
		interfaceType = interfaceType.Elem()
	}
	interfaceID := getQualifiedClassName(interfaceType)
	if interfaceType.Kind() != reflect.Interface {
		panic(fmt.Sprintf("Input type %v is not an interface", interfaceID))
	}
	if _, ok := r.iMap[interfaceID]; ok {
		panic(fmt.Sprintf("Interface %v is already bind", interfaceID))
	}
	result := &interfaceMetaInfo{}

	primaryType := reflect.TypeOf(primaryStruct)
	if !primaryType.Implements(interfaceType) {
		panic(fmt.Sprintf("The primary struct type %v not implement the interface %v", getQualifiedClassName(primaryType), interfaceID))
	}
	result.primary = getQualifiedClassName(primaryType)

	otherMap := make(map[string]bool)
	for _, each := range others {
		otherType := reflect.TypeOf(each)
		otherID := getQualifiedClassName(otherType)
		if !otherType.Implements(interfaceType) {
			panic(fmt.Sprintf("The struct type %v not implement the interface %v", otherID, interfaceID))
		}
		if _, ok := otherMap[otherID]; ok {
			panic(fmt.Sprintf("Can not bind %v twice on the same interface", otherID))
		}
		otherMap[otherID] = true
	}
	if _, ok := otherMap[result.primary]; ok {
		panic(fmt.Sprintf("Can not bind %v twice on the same interface", result.primary))
	}
	otherMap[result.primary] = true

	result.others = otherMap
	r.iMap[interfaceID] = result
}

func (r *register) ScanAndCheck() {
	// 1. Scan and Check interface
	for _, eachInterface := range r.iMap {
		eachInterface.nameMapping = make(map[string]string)
		for eachQualifier := range eachInterface.others {
			if meta, ok := r.sMap[eachQualifier]; ok {
				if name, exist := eachInterface.nameMapping[meta.option.name]; exist {
					panic(fmt.Sprintf("%v and %v share the same name %v for interface bind", eachQualifier, name, meta.option.name))
				}
				eachInterface.nameMapping[meta.option.name] = eachQualifier
			} else {
				panic(fmt.Sprintf("%v not found registered in vial", eachQualifier))
			}
		}
	}

	// 2. scan struct and find possible cycle injection
	// 0 - not start, 1 - pending, 2 - done
	checkMap := make(map[string]int)
	checkList := list.New()
	for name := range r.sMap {
		result := r.cycleInjectionCheck(name, checkMap, checkList)
		if !result {
			panic(printCycleInjectionLoop(name, checkList))
		}
	}
}

func (r *register) cycleInjectionCheck(name string, checkMap map[string]int, checkList *list.List) bool {
	if checkMap[name] == 2 {
		return true
	} else if checkMap[name] == 1 {
		return false
	}
	checkMap[name] = 1
	metaInfo := r.sMap[name]
	if metaInfo == nil {
		panic(fmt.Sprintf("Not found %v registered in the Vial", name))
	}
	for _, info := range metaInfo.dependency {
		nextName := info.name
		checkName := info.name
		if info.kind == interfaceKind {
			bindInfo, exist := r.iMap[nextName]
			if !exist {
				panic(fmt.Sprintf("not find bind information for interface %v", nextName))
			}
			if info.qualifier != "" {
				if mapping, found := bindInfo.nameMapping[info.qualifier]; found {
					info.reference = mapping
				} else {
					panic(fmt.Sprintf("qualifier %v not found for interface type %v bind", info.qualifier, info.name))
				}
			} else {
				info.reference = bindInfo.primary
			}
			checkName = fmt.Sprintf("%v(%v)", info.name, info.reference)
			nextName = info.reference
		} else if info.kind == valueKind {
			continue
		}

		el := checkList.PushBack(checkName)
		result := r.cycleInjectionCheck(nextName, checkMap, checkList)
		if !result {
			return false
		}
		checkList.Remove(el)
	}
	checkMap[name] = 2
	return true
}
