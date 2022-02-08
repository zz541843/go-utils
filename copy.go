package jz

import (
	"errors"
	"fmt"
	"reflect"
)

type jzCopy struct {
	ComplexSkip    bool
	HandlerFuncMap map[string]HandlerFunc
}
type HandlerFunc func(interface{}) (result interface{}, err error)

func NewCopy() *jzCopy {
	return &jzCopy{
		ComplexSkip:    true,
		HandlerFuncMap: map[string]HandlerFunc{},
	}
}
func (c *jzCopy) SetHandlerFuncMap(key string, handler HandlerFunc) {
	c.HandlerFuncMap[key] = handler
}
func (c *jzCopy) Struct2Map(src interface{}) (resMap map[string]interface{}) {
	resMap = make(map[string]interface{}, 1<<4)
	numSrcField := reflect.TypeOf(src).NumField()
	for index := 0; index < numSrcField; index++ {
		if !reflect.TypeOf(src).Field(index).IsExported() {
			continue
		}
		fieldName := reflect.TypeOf(src).Field(index).Name
		fieldValue := reflect.ValueOf(src).Field(index).Interface()
		// 结构体 递归
		if reflect.TypeOf(src).Field(index).Type.Kind() == reflect.Struct {
			// TODO 匿名结构体 暂时不管
			if !reflect.TypeOf(src).Field(index).Anonymous {
				fieldValue = c.Struct2Map(reflect.ValueOf(src).Field(index).Interface())
			}

		} else if reflect.TypeOf(src).Field(index).Type.Kind() == reflect.Array ||
			reflect.TypeOf(src).Field(index).Type.Kind() == reflect.Slice {
			if reflect.TypeOf(src).Field(index).Type.Elem().Kind() == reflect.Struct {
				// 结构体数组
				var arr []map[string]interface{}
				for sonIndex := 0; sonIndex < reflect.ValueOf(src).Field(index).Len(); sonIndex++ {
					sonMap := c.Struct2Map(reflect.ValueOf(src).Field(index).Index(sonIndex).Interface())
					arr = append(arr, sonMap)
				}
				fieldValue = arr
			} else {
				// 普通数组
				//fieldValue = reflect.ValueOf(src).Field(index).Interface()
			}
		}
		resMap[fieldName] = fieldValue
	}
	return
}

// StructCopy 相同字段结构体拷贝
// tar 目标指针，src 源
func (c *jzCopy) StructCopy(tar interface{}, src interface{}) (err error) {

	//类型判定
	// tar 必须是指针
	if reflect.TypeOf(tar).Kind() != reflect.Ptr && reflect.TypeOf(tar).Elem().Kind() != reflect.Struct {
		err = errors.New("tar value not a struct pointer")
		return
	}
	// src 可以为指针
	if reflect.TypeOf(src).Kind() == reflect.Ptr {
		if reflect.TypeOf(src).Elem().Kind() != reflect.Struct {
			err = errors.New("src value not a struct")
			return
		} else {
			//使用src构造源map
			err = c.Map2Struct(tar, c.Struct2Map(reflect.ValueOf(src).Elem().Interface()))
		}
	} else {
		if reflect.TypeOf(src).Kind() != reflect.Struct {
			err = errors.New("src value not a struct")
			return
		} else {
			//使用src构造源map
			err = c.Map2Struct(tar, c.Struct2Map(src))
			return
		}
	}

	if err != nil {
		return err
	}

	return
}

// Map2Struct map转结构体
// tar 必须是结构体指针
func (c *jzCopy) Map2Struct(tar interface{}, srcMap map[string]interface{}) (err error) {
	if reflect.TypeOf(tar).Kind() != reflect.Ptr && reflect.TypeOf(tar).Elem().Kind() != reflect.Struct {
		err = errors.New("tar value not a struct pointer")
		return
	}
	numDstField := reflect.TypeOf(tar).Elem().NumField()
	for index := 0; index < numDstField; index++ {
		tarFieldType := reflect.TypeOf(tar).Elem().Field(index)
		tarFieldTypeName := tarFieldType.Name
		tarFieldValue := reflect.ValueOf(tar).Elem().Field(index)
		if srcMap[tarFieldTypeName] == nil {
			continue
		}

		srcMapCurrentValue := reflect.ValueOf(srcMap[tarFieldTypeName])
		// 当前字段是结构体,递归赋值
		if tarFieldValue.Kind() == reflect.Struct {
			if reflect.ValueOf(srcMap[tarFieldTypeName]).Kind() != reflect.Map {
				continue
			}
			newStruct := reflect.New(tarFieldType.Type)
			err := c.Map2Struct(newStruct.Interface(), srcMap[tarFieldTypeName].(map[string]interface{}))
			if err != nil {
				continue
			}
			tarFieldValue.Set(newStruct.Elem())
			continue
		} else if tarFieldType.Type.Kind() == reflect.Array || tarFieldType.Type.Kind() == reflect.Slice {
			if tarFieldType.Type.Elem().Kind() == reflect.Struct {
				// 结构体数组
				currentStructArray := srcMap[tarFieldTypeName].([]map[string]interface{})
				newReflectArray := reflect.MakeSlice(tarFieldType.Type, len(currentStructArray), len(currentStructArray))
				for i := 0; i < len(currentStructArray); i++ {
					newStruct := reflect.New(tarFieldType.Type.Elem())
					err := c.Map2Struct(newStruct.Interface(), currentStructArray[i])
					if err != nil {
						continue
					}
					newReflectArray.Index(i).Set(newStruct.Elem())
				}
				tarFieldValue.Set(newReflectArray)
				continue
			} else {
				// 普通数组

			}
		}
		// 如果在这里，初步认定， 是基础类型
		// 如果对基础类型做了封装，可以进行赋值，因为封装了还是基础类型， 所以要使用kind
		// 类型不一致,则不赋值

		if srcMapCurrentValue.Type().Kind() != tarFieldValue.Type().Kind() {
			continue
		}

		// TODO 限制 tar基本字段必须是基本类型，复杂类型不支持？？？

		// 此时，数组也会在这里，但不需要管，因为默认能直接互相转换
		var newSrc interface{}
		if srcMapCurrentValue.Type() != tarFieldValue.Type() {
			// 进到这里，要么是tar是复杂类型，要么是src是复杂类型
			newSrc = srcMapCurrentValue.Interface()
			if IsBasicType(newSrc) {
				newTar := reflect.New(tarFieldType.Type)
				switch srcMapCurrentValue.Type().Kind() {
				case reflect.String:
					newTar.Elem().SetString(srcMapCurrentValue.String())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					newTar.Elem().SetInt(srcMapCurrentValue.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					newTar.Elem().SetUint(srcMapCurrentValue.Uint())
				}
				newSrc = newTar.Elem().Interface()
			} else if IsBasicType(tarFieldValue.Interface()) {
				switch srcMapCurrentValue.Type().Kind() {
				case reflect.String:
					newSrc = srcMapCurrentValue.String()
				case reflect.Int:
					newSrc = srcMapCurrentValue.Int()
				case reflect.Int8:
					newSrc = int8(srcMapCurrentValue.Int())
				case reflect.Int16:
					newSrc = int16(srcMapCurrentValue.Int())
				case reflect.Int32:
					newSrc = int32(srcMapCurrentValue.Int())
				case reflect.Int64:
					newSrc = int64(srcMapCurrentValue.Int())
				case reflect.Uint:
					newSrc = uint(srcMapCurrentValue.Uint())
				case reflect.Uint8:
					newSrc = uint8(srcMapCurrentValue.Uint())
				case reflect.Uint16:
					newSrc = uint16(srcMapCurrentValue.Uint())
				case reflect.Uint32:
					newSrc = uint32(srcMapCurrentValue.Uint())
				case reflect.Uint64:
					newSrc = uint64(srcMapCurrentValue.Uint())
				}
			} else {
				return fmt.Errorf("居然有第三种情况？？")
			}

			// 第二种方案 实现Scanner 和Valuer
			/*fmt.Println(srcMapCurrentValue.Type())
			if srcMapCurrentValue.Type().Implements(reflect.TypeOf((*JzCopyValue)(nil)).Elem()) {
				newVal, _ = srcMapCurrentValue.Interface().(JzCopyValue).JzValue()
			}
			fmt.Println(reflect.New(tarFieldValue.Type()).Type())
			//fmt.Println(reflect.New(tarFieldValue.Type()).Type().Elem())
			if reflect.New(tarFieldValue.Type()).Type().Implements(reflect.TypeOf((*JzCopyScan)(nil)).Elem()) {
				err := reflect.New(tarFieldValue.Type()).Elem().Interface().(JzCopyScan).JzScan(srcMapCurrentValue.Interface())
				if err != nil {
					return err
				}
			}*/

			/*// 不处理复杂类型转换
			if c.ComplexSkip {
				continue
			}
			// 外部自定义处理类型函数
			if len(c.HandlerFuncMap) > 0 {
				if handler, has := c.HandlerFuncMap[srcMapCurrentValue.Type().String()]; has {
					result, err := handler(srcMapCurrentValue.Interface())
					if err != nil {
						return err
					}
					tarFieldValue.Set(reflect.ValueOf(result))
					continue
				}
				return fmt.Errorf(srcMapCurrentValue.Type().String() + "没有对应的HandlerFunc")
			}*/
		}
		//tarFieldValue.Set(reflect.ValueOf(newVal))
		/*var setValue reflect.Value
		 */
		tarFieldValue.Set(reflect.ValueOf(newSrc))
	}
	return
}

type JzCopyScan interface {
	JzScan(value interface{}) error
}
type JzCopyValue interface {
	JzValue() (value interface{}, err error)
}

func IsBasicType(in interface{}) (b bool) {
	switch in.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		return true
	default:
		return false
	}
}
