package vial

func RegisterStruct[T any](options ...applyOption) {
	RegisterStructToContainer[T](c, options...)
}

func RegisterStructToContainer[T any](ctr *Container, options ...applyOption) {
	var structType T
	ctr.RegisterStructByInstance(structType, options...)
}

func RegisterStructByInstance(structType interface{}, options ...applyOption) {
	c.RegisterStructByInstance(structType, options...)
}

func RegisterConstructor(constructor interface{}, options ...applyOption) {
	c.RegisterConstructor(constructor, options...)
}

func Bind[T any, P any](others ...interface{}) {
	BindToContainer[T, P](c, others...)
}

func BindToContainer[T any, P any](ctr *Container, others ...interface{}) {
	i := new(T)
	var primaryStruct P
	ctr.Bind(i, primaryStruct, others...)
}

func BindByInstance(i interface{}, primaryStruct interface{}, others ...interface{}) {
	c.Bind(i, primaryStruct, others...)
}

func GetByInstance(dataType interface{}) (interface{}, error) {
	return c.GetByInstance(dataType)
}

func Get[T any]() (T, error) {
	return GetFromContainer[T](c)
}

func GetFromContainer[T any](ctr *Container) (T, error) {
	var data T
	value, err := ctr.GetByInstance(data)
	if err != nil {
		return data, err
	}
	return value.(T), err
}

func Done() {
	c.Done()
}

func NewContainer() *Container {
	return newContainer()
}
