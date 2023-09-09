# Vial - GoLang Spring Like Dependency Injection (DI) Component
Vial is a dependency injection (Spring pattern) component designed to be used by GoLang. Vial is inspired by Spring and google Wire. The purpose of Vial is to break the limitations of Wire and provide a more flexible dependency injection for GoLang.

As a new project, we do need your support. You are always welcome to communicate with us about the any Vial's problems. We are very happy to accept any features, issues, bug reports, etc., which are very important to us, and we also welcome joining the open source.
## Why Vial not Wire

The original intention of designing this component was that when our project using wire, I felt some limitations: 

1. There is no singleton mode. For example, when I want to write two init functions which can return two services, and both services depend on the same database instance, there will be a limitation. If I write them separately, then I will generate two duplicate database instances. If I put them together within one init function, I have to create a Service Collection to wrap them.
2. The original intention of interface design is to have many different implementation structs. In wire, the interface can only bind one struct. This results in having many aliases for the interface, or simply abandon the use of the interface.
3. When my service depends on multiple fields with the same interface but different implementations, I have to create an InterfaceCollection struct to collect all implementations and then select different implementations in the constructor.
4. When there is a problem with our injection, Wire sometimes cannot give us a clear message about where the problem lies. Sometimes there will even be a prompt similar to "cannot find bindings for *invalid type".

Based on these, I developed the Vial v0.0.1 version, hoping to provide a Spring like methods for users with more convenience. At the same time, I also borrowed some wire syntax.

## Features of Vial

1. Unlike Wire, Vial is a runtime DI component.
2. Vial currently supports two types of Scope, Prototype and Singleton. All struct injections default to Singleton.
3. Like wire, Vial supports Struct-based injection, constructor-based injection and interface-based binding.
4. When injecting Struct, you can choose automatic injection or constant injection for each exported field, saving you time in writing constructors.
5. Unlike Wire, our Vial supports binding multiple struct implementations to an interface and selecting a primary implementation. When injecting an interface, we inject the primary struct first, unless the user declares other structs through the tag "qualifier"
6. Vial uses generic type in its syntax, which allows you to inject without declaring a variable, and you can better specify whether you want StructA, *StructA, or even ***StructA.At the same time, this also provides some convenience for you when using IDE. You do not need to convert the interface to StructA, and you can also get the syntax support of the IDE.(However, since generic type was introduced in version 1.18 and later, this also results in the current Vial needing to be used in version 1.18 or above. If you are sure that the project cannot be upgraded to version 1.18 and above, you can contact us and we can consider removing this feature and provide a version more like the original version of wire for you to use.)
7. Vial provides a solution for creating multiple containers, which means you can create multiple independent dependency injection environments or maintain multiple singleton pools.
8. Vial provides a very rich detection mechanism, such as whether there is a lack of dependencies, whether there are instances with duplicate names, whether there are circular dependencies, etc., and provides you with detailed information through panic when the system starts. This ensures that all problems can be discovered before the program is started, rather than discovered during operation.

## Usage

`go get "github.com/GarrickZ2/vial"`

### Inject By Struct

````golang
import	"github.com/GarrickZ2/vial"
type StructA struct {
  FieldB StructB   `auto_wire:""`
  FieldC *StructC  `auto_wire:""`
  ValueA int32     `value:"21"`
  valueB byte
  fieldD StructD
}

// Vial-like method
func init() {
  vial.RegisterStruct[StructA]()
  vial.RegisterStruct[StructB](vial.WithProtoType())
  vial.RegisterStruct[*StructC](vial.WithSingleton(), vial.WithName("BeanName"))
  vial.Done()  // Has to use vial.Done at the end to indicate the initialization finished
}

// Wire-like method
func init() {
  vial.RegisterStructByInstance(StructA{})
  vial.RegisterStructByInstance(StructB{}, vial.WithProtoType())
  vial.RegisterStructByInstance(new(StructC), vial.WithSingleton())
  vial.Done()
}
````

1.   We support two method to do the injection for Struct. a) Vial-like method `vial.RegisterStruct[T any](options...)` b) Wire-like method `vial.RegisterStructByInstance(data interface{}, options...)`
2.   The default scope with injected Struct is `Singleton`, you can use options to change this setting.
3.   You can use `vial.WithName(name string)` to define a bean name for a struct, this might be useful for later Binding. We will use the Struct Name as the default bean name.
4.   Within a struct, we provide several tags to use help the injection. 
     1.   `auto_wire`: means you hope this filed get injected. 
     2.   `value`: can help you set a default value to an original data type besides `chan` ,`uintptr`, `array` `slice`, `struct` and `map`. If will validate whether the value can be converted into the correct data type, if not, we will panic at init time.
     3.   `qualifier`: When you want to use a non-primary struct for interface injection, you can use qualifier to specify a bean name.
     4.   ... welcome any suggestions for more useful tags
5.   For the same container, Vial cannot accept register same type struct. `Same` is defined by FullQualifiedName, `StructA` , `*StructA` and `**StructA` are different types.



### Inject By Constructor

````golang
func NewStructA(data StructB, data2 *StructC) *StructA {
  ...
}

func NewStructB() (StructB, error) {
  ...
}

func init() {
  vial.RegisterConstructor(NewStructA)
  vial.RegisterConstructor(NewStructB, vial.WithProtoType())
  vial.RegisterStruct[*NewStructC]()
  
  vial.Done()
}
````

1.   `RegisterConstructor(constructor interface{}, options...)` This method is quite similar with Wire's one.
2.   You can provide a constructor with 1 or 2 return data. For 1 return params constructor, it has to be the type you want to register. For 2 return params constructor, it has to be the registered type with an error. If error happened, we will give you the error when creating the exec the constructor.
3.   Options are similar to InjectByStruct, you can define a struct's scope and bean name.

### Interface Binding

````go
type TestInterface interface{
  
}

type StructA {
  FieldB TestInterface `auto_wire:""`                       // *StructB
  FieldC TestInterface `auto_wire:"" qualifier:"StructC"`   // StructC
  FieldD TestInterface `auto_wire:"" qualifier:"TestName"`  // *StructD
}

func init() {
  vial.RegisterStruct[*StructB]()
  vial.RegisterStruct[StructC]()
  vial.RegisterStruct[*StructD](vial.WithName("TestName"))
  
  vial.Bind[TestInterface, *StructB](StructC{}, new(StructD)) // Vial-like method
  vial.Bind(new(TestInterface), new(StructB), StructC{}, new(StructD)) // Wire-like method
  vial.Done()
}
````

1.   We provide a vial-like method `Bind[interfaceType, primaryStructType](otherStructTypes...)` and a wire-like method `Bind(interfaceType, primaryStructType, otherStructTypes...)`.
2.   You can provide more than one struct to bind on one interface. There will be one primary struct and others struct. For the interface injection, if there is no qualifier tag, we will directly use the primary struct. With the qualifier tag, we will find the others binding struct with the correct bean name.
3.   During binding, we will ensure each struct implements the interface, or there will be a panic.
4.   If there is a conflict bean name, we will throw a panic. Please use `vial.WithName()` option to assign another name.
5.   If the qualifier points to a non-binding struct or non-exist struct, we will throw a panic during init.

### If you'd like a provider method in Wire

````go
// Package A
func ProviderA() {
  vial.RegisterStruct[StructA]()
  // ...
  vial.Bind[InterfaceA, StructA]()
}

// Package B
func ProviderB() {
  vial.RegisterStruct[StructB]()
  // ...
  vial.RegisterConstructor(NewStructC, vial.WithProtoType())
}

// main
func init() {
  ProviderA()
  ProviderB()
  vial.RegisterStruct[Other]()
  vial.Done()
}
````

1.   Inject Order (Sequence) is not important in vial
2.   Please only use and must use `vial.Done()` in main.init().
3.   Before initialization finished (before `vial.Done()` is called), please only allow one main go-routine to operate on the Vial. (We don't provide any multi-thread protection in init phase). But we do support a perfect multi-thread access after initialization (after `vial.Done()`)

### Get Value From Vial

````go
func main() {
  // after vial.Done() in init() func
  
  // vial like get method
  structA, err := vial.Get[*StructA]()
 
  // you can directly access structA's fields
  fmt.Println(structA.Data, structA.Children[0].Age)
  
  // wire like get method
  structAI, err := vial.GetByInstance(new(StructA))
  structA := structAI.(*StructA)
  fmt.Println(structA.Data, structA.Children[0].Age)
}
````

1.   We provide two method to get value from vial. Vial-like method `vial.Get[T any]()(T, error)` and Wire-like method `vial.GetByInstance(StructTypeInstance)(interface{}, error)`
2.   Both methods return one instance and one error. The error comes from your constructor's return error.
3.   Vial-like method can directly return the data type you required, you can use them directly. However, the Wire-like method will return an interface, you need one further step to do the type conversion.



### Multiple Containers

````go
var container1 *vial.Container
var container2 *vial.Container

func init() {
  container1 = vial.NewContainer()
  container2 = vial.NewContainer()
  
  vial.RegisterStruct[StructA]()
  
  container1.RegisterStructByInstance(StructA{})
  container1.RegisterConstructor(NewStructB)
  vial.RegisterStructToContainer[StructA](container1, options...)
  
  vial.Done()
  container1.Done()
  container2.Done()
}

func main() {
  structA, err := container2.GetByInstance(new(StructA))
  // or
  structA, err := vial.GetFromContainer[StructA](container2)
}
````

1.   The package vial contains a primary container. The container between `vial`, `container1` and `container2` are completely seperated. You can use any one of them independently.
2.   The generated container doesn't contain any generic type method, i.e. `container.RegisterStruct[T any](options...)` , `container.Bind[Interface any, PrimaryStruct any](others...)`  and `container.Get[T any]() (T error)`. The generated container only contains the `ByInstance` method.
3.   To access the generic method for generated container, please use `vial.RegisterStructToContainer[T any](c *vial.Container, options...)`, `vial.BindToContainer[T any, P any](c *vial.Container, others...)` and `vial.GetFromContainer[T any](c *vial.Container) (T, error)` instead.

## At Last

We will provide more tutorial docs and example codes in the future.

The project is still very new and may not be mature enough or powerful enough, so please don't be stingy with your suggestions. We look forward to making Vial stronger together. Welcome to use Vial to help you make dependency injection easier in golang.
