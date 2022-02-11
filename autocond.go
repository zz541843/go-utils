package jz

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

type gormBuild struct {
	db *gorm.DB
}

//AutoBuildGormCondition 构建 cond
func AutoBuildGormCondition(db *gorm.DB, tar interface{}) (err error) {
	if reflect.TypeOf(tar).Kind() != reflect.Struct {
		err = errors.New("tar value not a struct pointer")
		return
	}
	g := &gormBuild{
		db: db,
	}
	if reflect.TypeOf(tar).Kind() == reflect.Ptr {
		g.generateCond(db, reflect.ValueOf(tar).Elem().Interface())
	} else {
		g.generateCond(db, tar)
	}
	return nil
}

//generateCond 生成cond
func (g *gormBuild) generateCond(db *gorm.DB, str interface{}) {
	v := reflect.ValueOf(str)
	t := reflect.TypeOf(str)
	fieldNum := t.NumField()
	if fieldNum == 0 {
		return
	}
	for i := 0; i < fieldNum; i++ {
		typeField := t.Field(i)
		valueField := v.Field(i)
		condTag := typeField.Tag.Get("jz-cond")
		if len(condTag) == 0 {
			continue
		}
		tags := strings.Split(condTag, ",")
		cond, col, value := g.checkCond(cond{
			fieldName:  typeField.Name,
			valueField: valueField,
			tags:       tags,
		})
		g.caseCond(db, cond, col, value)
	}
	return
}

type cond struct {
	fieldName  string
	valueField reflect.Value
	tags       []string
}

func SnakeCaseFromString(str string) string {
	var key string
	for k := 0; k < len(str); k++ {
		if str[k] >= 65 && str[k] <= 90 {
			if k == 0 {
				key = string(str[k] + 32)
			} else {
				key += "_" + string(str[k]+32)
			}
		} else {
			key += string(str[k])
		}
	}
	return key
}

//checkCond 检查是否符合条件
func (g *gormBuild) checkCond(req cond) (cond, col string, value interface{}) {
	if req.valueField.Kind() == reflect.Ptr {
		value = req.valueField.Elem().Interface()
	} else {
		value = req.valueField.Interface()
	}

	if len(req.tags) == 1 {
		cond = req.tags[0]
	} else if len(req.tags) == 2 {
		col = req.tags[1]
	}
	if len(col) == 0 {
		if cond == "in" && strings.LastIndex(req.fieldName, "s") == len(req.fieldName)-1 {
			col = req.fieldName[:len(req.fieldName)-1]
		} else {
			col = SnakeCaseFromString(req.fieldName)
		}
	}

	// 默认给in条件的 表字段名称设置为结构体字段去s，没有s就一模一样

	return
}
func (g *gormBuild) caseCond(db *gorm.DB, cond string, col string, val interface{}) {

	cond = strings.ToLower(cond)
	switch cond {
	case "eq":
		db.Where(fmt.Sprintf("%s = ?", col), val)
	case "neq":
		db.Where(fmt.Sprintf("%s != ?", col), val)
	case "like":
		db.Where(fmt.Sprintf("%s like ?", col), val)
	case "null":
		db.Where(fmt.Sprintf("%s is null", col), val)
	case "notnull":
		db.Where(fmt.Sprintf("%s is not null", col), val)
	case "in":
		db.Where(fmt.Sprintf("%s in ?", col), val)
	case "notin":
		db.Not(map[string]interface{}{col: val})
	case "gt":
		db.Where(fmt.Sprintf("%s > ?", col), val)
	case "lt":
		db.Where(fmt.Sprintf("%s < ?", col), val)
	case "gte":
		db.Where(fmt.Sprintf("%s >= ?", col), val)
	case "lte":
		db.Where(fmt.Sprintf("%s <= ?", col), val)
	case "sql":
	default:
		db.Where(fmt.Sprintf("%s = ?", col), val)
	}
}
