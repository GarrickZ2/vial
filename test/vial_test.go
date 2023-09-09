package test

import (
	"fmt"
	"testing"

	"github.com/GarrickZ2/vial"
)

type Agile interface {
}

type StructB struct {
	B Agile `auto_wire:""`
}

type TestStruct struct {
	a int
	b byte
	C StructB `auto_wire:""`
}

type StructC struct {
	Number int32 `auto_wire:"" value:"29"`
}

type StructD struct {
	Data  string  `value:"data value ok"`
	Data2 float32 `value:"13.2"`
}

type Data int

func NewData() Data {
	return 9
}

func TestVial(t *testing.T) {
	vial.RegisterStruct[TestStruct](vial.WithProtoType())
	vial.RegisterStruct[StructB](vial.WithProtoType())
	vial.RegisterStruct[*StructC](vial.WithSingleton())
	vial.RegisterStruct[StructD]()
	vial.RegisterConstructor(NewData)
	vial.Bind[Agile, *StructC](StructD{}, TestStruct{})

	vial.Done()

	b, _ := vial.Get[StructB]()
	fmt.Println(b.B)
}
