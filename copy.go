package jz

import (
	"errors"
	"fmt"
	"reflect"
)

func Struct2Map(src interface{}) (resMap map[string]interface{}) {
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
				fieldValue = Struct2Map(reflect.ValueOf(src).Field(index).Interface())
			}

		} else if reflect.TypeOf(src).Field(index).Type.Kind() == reflect.Array ||
			reflect.TypeOf(src).Field(index).Type.Kind() == reflect.Slice {
			if reflect.TypeOf(src).Field(index).Type.Elem().Kind() == reflect.Struct {
				// 结构体数组
				var arr []map[string]interface{}
				for sonIndex := 0; sonIndex < reflect.ValueOf(src).Field(index).Len(); sonIndex++ {
					sonMap := Struct2Map(reflect.ValueOf(src).Field(index).Index(sonIndex).Interface())
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
func StructCopy(tar interface{}, src interface{}) (err error) {

	//类型判定
	// tar 必须是指针
	if reflect.TypeOf(tar).Kind() != reflect.Ptr && reflect.TypeOf(tar).Elem().Kind() != reflect.Struct {
		err = errors.New("tar value not a struct pointer")
		return
	}
	// src 先限制为不是指针
	if reflect.TypeOf(src).Kind() != reflect.Struct {
		err = errors.New("src value not a struct")
		return
	}

	//使用src构造源map
	srcMap := Struct2Map(src)
	err = Map2Struct(tar, srcMap)
	if err != nil {
		return err
	}

	return
}

// Map2Struct map转结构体
// tar 必须是结构体指针
func Map2Struct(tar interface{}, srcMap map[string]interface{}) (err error) {
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
			err := Map2Struct(newStruct.Interface(), srcMap[tarFieldTypeName].(map[string]interface{}))
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
					fmt.Println(newStruct.Interface(), currentStructArray[i])
					err := Map2Struct(newStruct.Interface(), currentStructArray[i])
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
		// 类型不一致,则不赋值
		if srcMapCurrentValue.Type().Name() != tarFieldValue.Type().Name() {
			continue
		}
		tarFieldValue.Set(reflect.ValueOf(srcMapCurrentValue.Interface()))
	}
	return
}
