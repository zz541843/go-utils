package tests

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	jz "github.com/zz541843/go-utils"
	"testing"
	"time"
)

type myString string
type myStringArr []string
type A struct {
	DD
}
type DD struct {
	Name string
}
type B struct {
	DD
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
		DD: DD{
			Name: "1",
		},
	}
	jzCopy := jz.Copier{}
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

type Teacher struct {
	Name Mystring
	Time MyTime
}
type Student struct {
	Name string
	Time time.Time
}
type Mystring string
type MyTime time.Time

func TestB(t *testing.T) {
	tt := Teacher{}
	ss := Student{
		Name: "1",
		Time: time.Now(),
	}
	jzCopy := jz.Copier{}
	jzCopy.StructCopy(&tt, ss)

	spew.Dump(ss)
	spew.Dump(tt)
}
