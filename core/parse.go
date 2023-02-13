package core

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const defaulttimeformate = "2006-01-02 15:04:05"

var maptimeformate = map[*regexp.Regexp]string{
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}\s{1}\d{2}\:\d{2}:\d{2}$`):  "2006-01-02 15:04:05",
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T]{1}\d{2}\:\d{2}:\d{2}$`): "2006-01-02 15:04:05",
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}$`):                         "2006-01-02",
	regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s{1}\d{2}\:\d{2}:\d{2}$`):  "2006-01-02 15:04:05",
	regexp.MustCompile(`\d{4}/\d{2}/\d{2}[T]{1}\d{2}\:\d{2}:\d{2}$`): "2006-01-02 15:04:05",

	regexp.MustCompile(`\d{4}/\d{2}/\d{2}$`): "2006-01-02",
	//2023-02-08T16:02:00.000Z
}

func PredictTimeFormat(val string) string {
	timeFormat := ""
	bvar := []byte(val)
	for reg, fmt := range maptimeformate {
		if reg.Match(bvar) {
			timeFormat = fmt
		}
	}
	return timeFormat
}

// 智能化的bind,自动识别JSON、form
func Bind(req *http.Request, obj interface{}) error {
	contentType := req.Header.Get("Content-Type")
	//如果是简单的json
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		return BindJson(req, obj)
	} else if strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded") {
		return BindForm(req, obj)
	} else if strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
		return BindForm(req, obj)
	} else {
		return BindForm(req, obj)
	}
	// return errors.New("当前格式类型" + contentType + "暂不支持")
}

func BindJson(req *http.Request, obj interface{}) error {
	s, err := ioutil.ReadAll(req.Body) //把  body 内容读入字符串
	if err != nil {
		return err
	}
	if len(s) == 0 {
		return nil
	} else {
		err = json.Unmarshal(s, obj)
		return err
	}
}

func BindForm(req *http.Request, ptr interface{}) error {
	err := req.ParseForm()
	if err != nil {
		return err
	}
	//fmt.Println("BindForm ==>" + req.RequestURI + "==>" + req.Form.Encode())
	err = mapForm(ptr, req.Form, req.URL.Query())
	return err
}

// 解析成json
func ParseObject(input []byte, obj interface{}) error {
	return json.Unmarshal(input, obj)
}

// 支持多种
func mapForm(ptr interface{}, form map[string][]string, externFormArr ...map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	for _, externForm := range externFormArr {
		for field, value := range externForm {
			form[field] = value
		}
	}
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()

		inputFieldName := typeField.Tag.Get("form")
		if inputFieldName == "" {
			inputFieldName = typeField.Name

			// if "form" tag is nil, we inspect if the field is a struct.
			// this would not make sense for JSON parsing but it does for a form
			// since data is flatten
			if structFieldKind == reflect.Struct {
				err := mapForm(structField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}

		numElems := len(inputValue)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for i := 0; i < numElems; i++ {
				if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if _, isTime := structField.Interface().(time.Time); isTime {
				if err := setTimeField(inputValue[0], typeField, structField); err != nil {
					return err
				}
				continue
			}
			if _, isJsonDate := structField.Interface().(Date); isJsonDate {
				if err := setJsonDateField(inputValue[0], typeField, structField); err != nil {
					return err
				}
				continue
			}
			if _, isJsonTime := structField.Interface().(DateTime); isJsonTime {
				if err := setJsonTimeField(inputValue[0], typeField, structField); err != nil {
					return err
				}
				continue
			}
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
				return err
			}
		}
	}
	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return nil
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = PredictTimeFormat(val)
	}
	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}

func setJsonDateField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		//return errors.New("Blank time format")
		timeFormat = PredictTimeFormat(val)
		if timeFormat == "" {
			timeFormat = defaulttimeformate
		}
	}
	//log.Println(1)
	if val == "" {
		value.Set(reflect.ValueOf(DateFromTime(time.Time{})))
		return nil
	}
	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}
	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}
	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	r := DateFromTime(t)
	value.Set(reflect.ValueOf(r))
	return nil
}
func setJsonTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = defaulttimeformate
		timeFormat = PredictTimeFormat(val)
		if timeFormat == "" {
			timeFormat = defaulttimeformate
		}
	}

	if val == "" {
		value.Set(reflect.ValueOf(DateTimeFromTime(time.Time{})))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	r := DateTimeFromTime(t)
	value.Set(reflect.ValueOf(r))
	return nil
}

// Don't pass in pointers to bind to. Can lead to bugs. See:
// https://github.com/codegangsta/martini-contrib/issues/40
// https://github.com/codegangsta/martini-contrib/pull/34#issuecomment-29683659
func ensureNotPointer(obj interface{}) {
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		panic("Pointers are not accepted as binding models")
	}
}
