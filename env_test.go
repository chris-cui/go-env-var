package envvar

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
)

type Person struct {
	Name string  `env:"PERSON_NAME" required:"true"`
	Age  int     `env:"PERSON_AGE" required:"true" converter:"toAge"`
	Desc *string `env:"PERSON_DESC"`
	Car  *Car
}

type Car struct {
	Brand  string  `env:"CAR_BRAND" default:"Toyota"`
	Engine float64 `env:"CAR_ENGINE" required:"true" converter:"toFloat"`
	Model  string  `env:"CAR_MODEL"`
	Desc   *string `env:"CAR_DESC"`
}

func TestRequired(t *testing.T) {
	defer cleanup()

	p := &Person{}
	err := Load(p)
	assertTrue(t, err != nil, "should get error since Name is required")

	os.Setenv("PERSON_NAME", "Tom")
	p = &Person{}
	err = Load(p)
	assertTrue(t, err != nil, "should get error since Age is required")

	Converter("toAge", func(s string) (any, error) {
		return s, nil
	})
	os.Setenv("PERSON_AGE", "ten")
	p = &Person{}
	err = Load(p)
	assertTrue(t, err != nil, "should get error since converter returns incorrect type")

	ClearConverters()
	registerConverts()
	os.Setenv("PERSON_AGE", "ten")
	p = &Person{}
	err = Load(p)
	assertTrue(t, err != nil, "should get error since Age is integer")

	os.Setenv("PERSON_AGE", "7")
	p = &Person{}
	err = Load(p)
	assertTrue(t, err != nil, "should get error since Age is too small")

	os.Setenv("PERSON_AGE", "17")
	p = &Person{}
	err = Load(p)
	assertTrue(t, err == nil, "should not get error since all required fields are set")

	assertEqual(t, "Tom", p.Name, "should set Name")
	assertEqual(t, 17, p.Age, "should set Age")
	assertTrue(t, p.Desc == nil, "should not set Desc since no env variable")
}

func TestNested(t *testing.T) {
	defer cleanup()

	registerConverts()

	name := "Tom"
	age := 17
	personDesc := "Nice person"
	os.Setenv("PERSON_NAME", name)
	os.Setenv("PERSON_AGE", strconv.Itoa(age))
	os.Setenv("PERSON_DESC", personDesc)

	engine := 2.0
	model := "Camry"
	carDesc := "nothing to say"
	os.Setenv("CAR_ENGINE", fmt.Sprint(engine))
	os.Setenv("CAR_MODEL", model)
	os.Setenv("CAR_DESC", carDesc)

	p := &Person{
		Car: &Car{},
	}
	err := Load(p)
	assertTrue(t, err == nil, "should no error since all required fields are set")
	assertEqual(t, name, p.Name, "should set person name")
	assertEqual(t, age, p.Age, "should set person age")
	assertEqual(t, personDesc, *p.Desc, "should set person desc")

	c := p.Car
	assertEqual(t, "Toyota", c.Brand, "should set car brand from default")
	assertEqual(t, model, c.Model, "should set car model")
	assertEqual(t, engine, c.Engine, "should set car engine")
	assertEqual(t, carDesc, *c.Desc, "should set car desc")

}

func assertTrue(t *testing.T, b bool, msg ...string) {
	if !b {
		t.Errorf("expect true - %v", getMsg(msg...))
	}
}

func assertFalse(t *testing.T, b bool, msg ...string) {
	if b {
		t.Errorf("expect false - %v", getMsg(msg...))
	}
}

func assertEqual(t *testing.T, expected any, real any, msg ...string) {
	if expected != real {
		t.Errorf("expect %v but got %v - %v", expected, real, getMsg(msg...))
	}
}

func getMsg(msg ...string) string {
	if len(msg) > 0 {
		return msg[0]
	}
	return ""
}

func registerConverts() {
	Converter("toAge", func(s string) (any, error) {
		a, e := strconv.Atoi(s)
		if e != nil {
			return 0, e
		}
		if a < 10 {
			return 0, errors.New("age must be larger than 10")
		}
		return a, e
	})
	Converter("toFloat", func(s string) (any, error) {
		return strconv.ParseFloat(s, 64)
	})

}

func cleanup() {
	os.Clearenv()
	ClearConverters()
}
