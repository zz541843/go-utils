package tests

import (
	"fmt"
	jz "github.com/zz541843/go-utils"
	"reflect"
	"testing"
)

type myString string
type myStringArr []string
type A struct {
	Arr myString
}
type B struct {
	Arr string
}

func (a *myString) JzScan(value interface{}) error {
	v, flag := value.(string)
	if flag {
		*a = myString(v)
		return nil
	}
	return fmt.Errorf("转换类型不是string ！")
}

// Value return json value, implement driver.Valuer interface
func (a myString) JzValue() (interface{}, error) {
	return string(a), nil
}
func TestA(t *testing.T) {
	a := A{}
	b := B{
		Arr: "12",
	}
	jzCopy := jz.NewCopy()
	/*jzCopy.HandlerFuncMap["tests.myString"] = func(i interface{}) (result interface{}, err error) {
		str, flag := i.(myString)
		if flag {
			return string(str), nil
		} else {
			return nil, fmt.Errorf("类型错误！不是tests.myString！")
		}
	}
	jzCopy.HandlerFuncMap["tests.myStringArr"] = func(i interface{}) (result interface{}, err error) {
		str, flag := i.(myStringArr)
		if flag {
			return []string(str), nil
		} else {
			return nil, fmt.Errorf("类型错误！不是tests.myStringArr！")
		}
	}*/
	//jzCopy.ComplexSkip = false
	//err := jzCopy.StructCopy(&b, a)
	err := jzCopy.StructCopy(&a, b)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	jz.PrintStruct(a)
}

func TestB(t *testing.T) {
	var a int16
	a = 1
	n := reflect.New(reflect.TypeOf(a))
	n.Elem().SetInt(13766)
	fmt.Println(n.Elem())
	fmt.Println(n.Type())
}