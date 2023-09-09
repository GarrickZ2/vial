package vial

type scope int

const (
	singleton scope = iota
	protoType
)

const (
	autoWire  string = "auto_wire"
	qualifier string = "qualifier"
	value     string = "value"
)

type kindType int

const (
	valueKind kindType = iota
	structKind
	interfaceKind
)

type buildType int

const (
	buildByInject buildType = iota
	buildByConstructor
)
