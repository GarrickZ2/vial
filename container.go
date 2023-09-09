package vial

import "fmt"

var c *Container

type Container struct {
	initType   int
	register   *register
	collection *collection
}

func newContainer() *Container {
	return &Container{
		register:   newRegister(),
		collection: newCollection(),
	}
}

func (c *Container) buildSingletonMap() {
	singletonMap := make(map[string]*singletonEntry)
	for name, each := range c.register.sMap {
		if each.option.scope == singleton {
			metaInfo := each
			singletonMap[name] = &singletonEntry{
				metaInfo: metaInfo,
			}
		}
	}
	c.collection.singletonMap = singletonMap
}

func (c *Container) RegisterStructByInstance(structType interface{}, options ...applyOption) {
	if c.initType == 1 {
		panic("the vial has been initialized, cannot register more")
	}
	c.register.RegisterStruct(structType, options...)
}

func (c *Container) RegisterConstructor(constructor interface{}, options ...applyOption) {
	if c.initType == 1 {
		panic("the vial has been initialized, cannot register more")
	}
	c.register.RegisterConstruct(constructor, options...)
}

func (c *Container) Bind(i interface{}, primaryStruct interface{}, others ...interface{}) {
	if c.initType == 1 {
		panic("the vial has been initialized, cannot bind more")
	}
	c.register.Bind(i, primaryStruct, others...)
}

func (c *Container) Done() {
	if c.initType == 1 {
		panic("cannot call Done method twice")
	}
	c.register.ScanAndCheck()
	c.buildSingletonMap()
	c.initType = 1
}

func (c *Container) GetByInstance(dataType interface{}) (interface{}, error) {
	if c.initType != 1 {
		return nil, fmt.Errorf("vial hasn't been initialized")
	}
	return c.getValue(dataType)
}
