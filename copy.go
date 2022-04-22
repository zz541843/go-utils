package jz

import (
	"errors"
	"fmt"
	"reflect"
)

type Copier struct {
	Cover bool
	deep  int8
}
type HandlerFunc func(interface{}) (result interface{}, err error)

//func (c *Copier) SetHandlerFuncMap(key string, handler HandlerFunc) {
//	c.HandlerFuncMap[key] = handler
//}

// 匿名结构体默认匹配同名的匿名结构体中

// `jzcp:"-"` 不处理
// `jzcp:"<->"` 当作src子字段处理 仅在匿名结构体字段上有用
// `jzcp:"xxx"` src的该字段匹配tar的xxx字段

// 只有匿名结构体才能匹配匿名结构体
// 匿名结构体通过tag更改匹配对象，可匹配到tag的结构体

func (c *Copier) Struct2Map(src interface{}) (resMap map[string]interface{}) {
	c.deep += 1

	resMap = make(map[string]interface{}, 1<<4)
	numSrcField := reflect.TypeOf(src).NumField()
	for index := 0; index < numSrcField; index++ {
		currentField := reflect.TypeOf(src).Field(index)
		currentFieldValue := reflect.ValueOf(src).Field(index)
		if !currentField.IsExported() {
			continue
		}
		fieldName := currentField.Name
		fieldValue := currentFieldValue.Interface()
		jzcpTag := currentField.Tag.Get("jzcp")
		if jzcpTag == "-" {
			continue
		} else if jzcpTag != "" && jzcpTag != "<->" {
			fieldName = jzcpTag
		}
		// 结构体 递归
		if currentField.Type.Kind() == reflect.Struct {
			// TODO 匿名结构体 暂时不管
			if currentField.Anonymous {
				if jzcpTag == "<->" {
					anonymouscMap := c.Struct2Map(currentFieldValue.Interface())
					for k, v := range anonymouscMap {
						resMap[k] = v
					}
					continue
				} else if jzcpTag == "" {
					fieldName += "-Anonymous"
				}
			}
			fieldValue = c.Struct2Map(currentFieldValue.Interface())
		} else if currentField.Type.Kind() == reflect.Array ||
			currentField.Type.Kind() == reflect.Slice {
			if currentField.Type.Elem().Kind() == reflect.Struct {
				// 结构体数组
				var arr []map[string]interface{}
				for sonIndex := 0; sonIndex < currentFieldValue.Len(); sonIndex++ {
					sonMap := c.Struct2Map(currentFieldValue.Index(sonIndex).Interface())
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
func (c *Copier) StructCopy(tar interface{}, src interface{}) (err error) {

	defer func() {
		c.deep = 0
	}()
	if tar == nil {
		return errors.New("tar is nil")
	}
	if reflect.ValueOf(tar).IsNil() {
		return errors.New("tar is nil")
	}
	if src == nil {
		return errors.New("src is nil")
	}
	if reflect.ValueOf(src).IsNil() {
		return errors.New("src is nil")
	}

	//类型判定
	// tar 必须是指针
	if reflect.TypeOf(tar).Kind() != reflect.Ptr || reflect.TypeOf(tar).Elem().Kind() != reflect.Struct {
		err = errors.New("tar value not a struct pointer")
		return
	}
	// src 不能是指针 -- 规范
	if reflect.TypeOf(src).Kind() == reflect.Ptr {
		err = errors.New("src value can't is a pointer")
		return
	}
	if reflect.TypeOf(src).Kind() != reflect.Struct {
		err = errors.New("src value not a struct")
		return
	}

	err = c.Map2Struct(tar, c.Struct2Map(src))

	if err != nil {
		return err
	}

	return
}

// Map2Struct map转结构体
// tar 必须是结构体指针
func (c *Copier) Map2Struct(tar interface{}, srcMap map[string]interface{}) (err error) {
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
			if tarFieldType.Anonymous {
				// 如果tar的字段是匿名结构体，则要加后缀匹配src的map key
				tarFieldTypeName += "-Anonymous"
			}
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

		// 底层类型不一致 则直接跳过 说明不是一路人
		if srcMapCurrentValue.Type().Kind() != tarFieldValue.Type().Kind() {
			continue
		}

		// src为空值时 是否覆盖tar
		if srcMapCurrentValue.IsZero() && !c.Cover {
			continue
		}

		// 此时，数组也会在这里，但不需要管，因为默认能直接互相转换
		newSrc := srcMapCurrentValue.Interface()
		if srcMapCurrentValue.Type() != tarFieldValue.Type() {
			// 进到这里，要么是tar是复杂类型，要么是src是复杂类型 包装了一层的套皮类型
			if IsBasicType(newSrc) {
				newTar := reflect.New(tarFieldType.Type)
				switch srcMapCurrentValue.Type().Kind() {
				case reflect.String:
					newTar.Elem().SetString(srcMapCurrentValue.String())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					newTar.Elem().SetInt(srcMapCurrentValue.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					newTar.Elem().SetUint(srcMapCurrentValue.Uint())
				case reflect.Bool:
					newTar.Elem().SetBool(srcMapCurrentValue.Bool())
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
					newSrc = srcMapCurrentValue.Uint()
				case reflect.Bool:
					newSrc = srcMapCurrentValue.Bool()
				}
			} else {
				return fmt.Errorf("居然有第三种情况？？")
			}
			// 那么到这里，基本类型和其套皮类型之间的相互匹配已经完成了
			// 但是其他类型的套皮通过反射则不可能完成，介于有调用者特意为接口提实现接口来处理这个太过麻烦
			// 干脆就不处理，毕竟这种情况少之又少之少之又少
			// 遇到了就手写一下吧，头发快掉光了啊

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
