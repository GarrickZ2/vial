package vial

type option struct {
	scope scope
	name  string
}

func newDefaultOption() option {
	return option{scope: singleton}
}

type applyOption struct {
	apply func(config *option)
}

func WithSingleton() applyOption {
	return applyOption{func(config *option) {
		config.scope = singleton
	}}
}

func WithProtoType() applyOption {
	return applyOption{func(config *option) {
		config.scope = protoType
	}}
}

func WithName(name string) applyOption {
	return applyOption{func(config *option) {
		config.name = name
	}}
}
