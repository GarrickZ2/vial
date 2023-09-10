package test

import (
	"fmt"
	"github.com/GarrickZ2/vial"
	"testing"
)

type Agile interface {
}

type StructB struct {
	B Agile `auto_wire:"" qualifier:"StructD"`
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

func NewStructD(a Data) (StructD, error) {
	return StructD{
		Data:  "1",
		Data2: float32(a),
	}, nil
}

type Data int

func NewData() Data {
	return 9
}

func TestVial(t *testing.T) {
	vial.RegisterStruct[TestStruct](vial.WithProtoType())
	vial.RegisterStruct[StructB](vial.WithProtoType())
	vial.RegisterStruct[*StructC](vial.WithSingleton())
	vial.RegisterConstructor(NewData)
	vial.RegisterConstructor(NewStructD)
	vial.Bind[Agile, *StructC](StructD{}, TestStruct{})

	vial.Done()

	b, _ := vial.Get[StructB]()
	fmt.Println(b.B)
}
