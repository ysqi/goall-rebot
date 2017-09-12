package rebot

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-ini/ini"
)

// PageInfo 可从Issue comment中提炼出的信息。
type PageInfo struct {
	Title   string   `ini:"title" ck:"required"`
	URL     string   `ini:"url" ck:"required"`
	Tags    []string `ini:"tags"`
	Content string   `ini:"content"`
}

// ToIni 将Page内容转换为Ini
// 转换时只显示内容不为空的部分
func (p PageInfo) ToIni() string {
	var (
		b = bytes.Buffer{}
		t = reflect.TypeOf(p)
		v = reflect.ValueOf(&p).Elem()
	)
	for i := 0; i < t.NumField(); i++ {
		vField := v.Field(i)
		field := t.Field(i)
		iniStr := field.Tag.Get("ini")
		if iniStr == "-" {
			continue
		}
		// key := strings.Split(iniStr, ",")[0]
		key := iniStr
		var value string
		switch vField.Kind() {
		case reflect.Slice:
			if v1, ok := vField.Interface().([]string); !ok {
				panic(fmt.Errorf("unprocess %s", vField.Kind()))
			} else if len(v1) > 0 {
				value = strings.Join(v1, ",")
			}
		default:
			value = vField.String()
		}
		if value != "" {
			b.WriteString(key)
			b.WriteString(" = ")
			b.WriteString(value)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// NewPageInfo 根据section节点信息创建Page对象，如果内容不符合条件则返回错误信息
func NewPageInfo(comment string) (PageInfo, error) {
	p := PageInfo{}
	//检查
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return p, errors.New("comment不能为空")
	}
	//解析comment
	f, err := ini.InsensitiveLoad([]byte(comment))
	if err != nil {
		return p, fmt.Errorf("按ini格式解析comment失败,%s", err.Error())
	}
	var (
		key     string
		value   interface{}
		section = f.Section("") //默认取根级信息
		t       = reflect.TypeOf(p)
		v       = reflect.ValueOf(&p).Elem()
	)
	for i := 0; i < t.NumField(); i++ {
		vField := v.Field(i)
		if !vField.CanSet() {
			continue
		}

		field := t.Field(i)
		iniStr := field.Tag.Get("ini")
		if iniStr == "-" {
			continue
		}

		key = strings.Split(iniStr, ",")[0]

		//检查
		if field.Tag.Get("ck") == "required" {
			if !section.Haskey(key) {
				return p, fmt.Errorf("缺失%q项", key)
			}
		}
		//赋值
		value = (Key{section.Key(key)}).GetValue(field.Type)
		v.Field(i).Set(reflect.ValueOf(value))

	}
	return p, nil
}

// Key ini key 扩展
type Key struct {
	*ini.Key
}

// GetValue 获取数据值
func (k Key) GetValue(valueType reflect.Type) interface{} {
	// TODO: 暂时仅考虑[]string，string
	switch valueType.Kind() {
	case reflect.Slice:
		return k.GetStrings()
	}
	return k.Value()
}

// GetStrings 获取[]string,按中英文逗号和空格，以及# 进行分割处理
func (k Key) GetStrings() []string {
	str := k.Value()
	if str == "" {
		return []string{}
	}
	//按中英文逗号和空格，以及# 进行分割处理
	sep := ","
	r := strings.NewReplacer(
		"，", sep,
		" ", sep,
		"#", sep,
	)
	arr := strings.Split(r.Replace(str), sep)
	items := []string{}
	//去掉空项
	for _, t := range arr {
		t = strings.TrimSpace(t)
		if t != "" {
			items = append(items, t)
		}
	}
	return items
}
