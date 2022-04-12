package jsonobj

import (
	"bytes"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type JSONOBJ struct {
	data interface{}
}

//将字符串转为JSON对象,
//body 参数可以为 string  []byte  nil
func Unmarshal(body interface{}) (*JSONOBJ, error) {
	var data []byte

	if body == nil {
		data = []byte("{}")
	} else if s, ok := (body).(string); ok {
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			data = []byte("{}")
		} else {
			data = []byte(s)
		}
	} else if bin, ok := (body).([]byte); ok {
		s = strings.TrimSpace(string(bin))
		if len(s) == 0 {
			data = []byte("{}")
		} else {
			data = []byte(s)
		}
	} else {
		return nil, errors.New("unsupported type")
	}

	var obj interface{}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := d.Decode(&obj); err != nil {
		return nil, err
	}

	return &JSONOBJ{obj}, nil
}

//attach 到 json 对象
func Attach(obj interface{}) *JSONOBJ {
	if mp, ok := (obj).(map[string]interface{}); ok {
		return &JSONOBJ{mp}
	}

	if list, ok := (obj).([]interface{}); ok {
		return &JSONOBJ{list}
	}

	panic("Attach() object type assert failed")

	return &JSONOBJ{nil}
}

func (j *JSONOBJ) RawData() interface{} {
	return j.data
}

func (j *JSONOBJ) Marshal() ([]byte, error) {
	return json.Marshal(j.data)
}

func (j *JSONOBJ) Stringify() string {
	v, _ := json.Marshal(j.data)
	return string(v)
}

//read json from file, if file not exist, create new one,
//if content is empty, then return '{}'
func ReadJsonFromFile(fname string) (*JSONOBJ, error) {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, os.ModePerm|os.ModeTemporary)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bin, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(bin))
	if len(content) == 0 {
		content = "{}"
	}

	var obj interface{}
	d := json.NewDecoder(strings.NewReader(content))
	d.UseNumber()
	if err := d.Decode(&obj); err != nil {
		return nil, err
	}

	return &JSONOBJ{obj}, nil
}

// read json array from file, if file not exist, create new one,
// if content is empty, then return '[]'
func ReadArrayFromFile(fname string) (*JSONOBJ, error) {
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, os.ModePerm|os.ModeTemporary)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bin, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	content := strings.TrimSpace(string(bin))
	if len(content) == 0 {
		content = "[]"
	}

	var obj interface{}
	d := json.NewDecoder(strings.NewReader(content))
	d.UseNumber()
	if err := d.Decode(&obj); err != nil {
		return nil, err
	}

	if _, ok := obj.([]interface{}); !ok {
		return nil, errors.New("content is not json array")
	}

	return &JSONOBJ{obj}, nil
}

func ReadArrayFromString(szRaw string) (*JSONOBJ, error) {
	content := strings.TrimSpace(szRaw)
	if len(content) == 0 {
		content = "[]"
	}

	var obj interface{}
	d := json.NewDecoder(strings.NewReader(content))
	d.UseNumber()
	if err := d.Decode(&obj); err != nil {
		return nil, err
	}

	if _, ok := obj.([]interface{}); !ok {
		return nil, errors.New("content is not json array")
	}

	return &JSONOBJ{obj}, nil
}

//if the file not exist, create a new one,
//otherwise truncates it before writing
func (j *JSONOBJ) SaveToFile(fname string, wellFormat bool) error {
	var v []byte
	var err error

	if wellFormat {
		v, err = json.MarshalIndent(j.data, "", "")
	} else {
		v, err = json.Marshal(j.data)
	}

	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(fname), 0777)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, v, 0644)
}

func (j *JSONOBJ) Set(key string, v interface{}) error {
	m, err := j.Map()

	if err != nil {
		return err
	}

	m[key] = v

	return nil
}

func (j *JSONOBJ) Get(key string) *JSONOBJ {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
			return &JSONOBJ{val}
		}
	}

	return &JSONOBJ{nil}
}

func (j *JSONOBJ) Del(key string) *JSONOBJ {
	m, err := j.Map()
	if err == nil {
		delete(m, key)
	}

	return j
}

func (j *JSONOBJ) IsKeyExist(key string) bool {
	m, err := j.Map()
	if err == nil {
		if _, ok := m[key]; ok {
			return true
		}
	}

	return false
}

func (j *JSONOBJ) IsNull() bool {
	if j.data == nil {
		return true
	}

	return false
}

func (j *JSONOBJ) IsArray() bool {
	if _, ok := (j.data).([]interface{}); ok {
		return true
	}

	return false
}

func (j *JSONOBJ) IsMap() bool {
	if _, ok := (j.data).(map[string]interface{}); ok {
		return true
	}

	return false
}

func (j *JSONOBJ) NilToArray() *JSONOBJ {
	if j.data == nil {
		j.data = make([]interface{}, 0)
		return j
	}

	return j
}

func (j *JSONOBJ) NilToMap() *JSONOBJ {
	if j.data == nil {
		j.data = make(map[string]interface{})
		return j
	}

	return j
}

func (j *JSONOBJ) GetAt(index int) *JSONOBJ {
	a, err := j.Array()
	if err == nil {
		if len(a) > index && index >= 0 {
			return Attach(a[index])
		}
	}
	return &JSONOBJ{nil}
}

//删除数组中一项
func (j *JSONOBJ) DelAt(index int) error {
	a, ok := (j.data).([]interface{})
	if !ok {
		return errors.New("type assertion to []interface{} failed")
	}

	if index < 0 || index >= len(a) {
		return errors.New("out of index")
	}

	newArray := make([]interface{}, 0, len(a)-1)
	for i, v := range a {
		if i != index {
			newArray = append(newArray, v)
		}
	}

	j.data = newArray

	return nil
}

//数组尾部追加
func (j *JSONOBJ) PushBack(item interface{}) error {
	a, ok := (j.data).([]interface{})
	if !ok {
		return errors.New("type assertion to []interface{} failed")
	}

	a = append(a, item)
	j.data = a

	return nil
}

//数组头部添加
func (j *JSONOBJ) PushFront(item interface{}) error {
	a, ok := (j.data).([]interface{})
	if !ok {
		return errors.New("type assertion to []interface{} failed")
	}

	newArray := make([]interface{}, 0, len(a)+1)
	newArray = append(newArray, item)
	newArray = append(newArray, a...)
	j.data = newArray
	return nil
}

//数组尾部删除
func (j *JSONOBJ) PopBack() error {
	a, ok := (j.data).([]interface{})
	if !ok {
		return errors.New("type assertion to []interface{} failed")
	}
	return j.DelAt(len(a) - 1)
}

//数组头部删除
func (j *JSONOBJ) PopFront() error {
	return j.DelAt(0)
}

func (j *JSONOBJ) Length() int {
	a, err := j.Array()
	if err == nil {
		return len(a)
	}
	return 0
}

func (j *JSONOBJ) Array() ([]interface{}, error) {
	if a, ok := (j.data).([]interface{}); ok {
		return a, nil
	}

	return nil, errors.New("type assertion to []interface{} failed")
}

func (j *JSONOBJ) MustArray() []interface{} {
	a, err := j.Array()
	if err == nil {
		return a
	}

	return make([]interface{}, 0)
}

func (j *JSONOBJ) ArrayShuffle() *JSONOBJ {
	if a, ok := (j.data).([]interface{}); ok {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := range a {
			j := r.Intn(i + 1)
			a[i], a[j] = a[j], a[i]
		}
	}

	return j
}

func (j *JSONOBJ) Map() (map[string]interface{}, error) {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

func (j *JSONOBJ) MustMap() map[string]interface{} {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m
	}

	return make(map[string]interface{})
}

// Bool type asserts to `bool`
func (j *JSONOBJ) Bool() (bool, error) {
	if s, ok := (j.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

func (j *JSONOBJ) String() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}

	return "", errors.New("type assertion to string failed")
}

//强迫对象转为字符串
func (j *JSONOBJ) ToString() string {
	if s, ok := (j.data).(string); ok {
		return s
	}

	if textMarl, ok := (j.data).(encoding.TextMarshaler); ok {
		if txt, err := textMarl.MarshalText(); err == nil {
			return string(txt)
		}
	}

	return fmt.Sprintf("%v", j.data)
}

// StringArray type asserts to an `array` of `string`
func (j *JSONOBJ) StringArray() ([]string, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}

	retArr := make([]string, 0, len(arr))

	for _, a := range arr {
		s, ok := a.(string)
		if !ok {
			return nil, errors.New("type assertion to string array failed")
		}

		retArr = append(retArr, s)
	}

	return retArr, nil
}

func (j *JSONOBJ) MapArray() ([]map[string]interface{}, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}

	retArr := make([]map[string]interface{}, 0, len(arr))

	for _, a := range arr {
		if a == nil {
			retArr = append(retArr, make(map[string]interface{}))
			continue
		}

		s, ok := a.(map[string]interface{})
		if !ok {
			return nil, errors.New("type assertion to []map[string]interface{} failed")
		}

		retArr = append(retArr, s)
	}

	return retArr, nil
}

func (j *JSONOBJ) MustMapArray() []map[string]interface{} {
	v, err := j.MapArray()
	if err == nil {
		return v
	} else {
		return make([]map[string]interface{}, 0)
	}
}

func (j *JSONOBJ) Int64Array() ([]int64, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}

	retArr := make([]int64, 0, len(arr))

	for _, a := range arr {
		item := &JSONOBJ{a}
		num, err := item.Int64()
		if err != nil {
			return nil, err
		}

		retArr = append(retArr, num)
	}

	return retArr, nil
}

func (j *JSONOBJ) Float64Array() ([]float64, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}

	retArr := make([]float64, 0, len(arr))

	for _, a := range arr {
		item := &JSONOBJ{a}
		num, err := item.Float64()
		if err != nil {
			return nil, err
		}

		retArr = append(retArr, num)
	}

	return retArr, nil
}

func (j *JSONOBJ) MustStringArray() []string {
	v, err := j.StringArray()
	if err == nil {
		return v
	} else {
		return make([]string, 0)
	}
}

func (j *JSONOBJ) MustString(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustString() received too many arguments %d", len(args)))
	}

	s, err := j.String()
	if err == nil {
		return s
	}

	return def
}

func (j *JSONOBJ) MustStringTrimSpace(args ...string) string {
	out := j.MustString(args...)
	out = strings.TrimSpace(out)
	return out
}

func (j *JSONOBJ) MustInt64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustInt() received too many arguments %d", len(args)))
	}

	i, err := j.Int64()
	if err == nil {
		return i
	}

	return def
}

func (j *JSONOBJ) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustInt() received too many arguments %d", len(args)))
	}

	i, err := j.Int64()
	if err == nil {
		return int(i)
	}

	return def
}

func (j *JSONOBJ) MustInt64Array() []int64 {
	v, err := j.Int64Array()
	if err == nil {
		return v
	} else {
		return make([]int64, 0)
	}
}

func (j *JSONOBJ) MustFloat64Array() []float64 {
	v, err := j.Float64Array()
	if err == nil {
		return v
	} else {
		return make([]float64, 0)
	}
}

func (j *JSONOBJ) MustFloat64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustFloat64() received too many arguments %d", len(args)))
	}

	f, err := j.Float64()
	if err == nil {
		return f
	}

	return def
}

func (j *JSONOBJ) MustBool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustBool() received too many arguments %d", len(args)))
	}

	b, err := j.Bool()
	if err == nil {
		return b
	}

	return def
}

func (j *JSONOBJ) MustUint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustUint64() received too many arguments %d", len(args)))
	}

	i, err := j.Uint64()
	if err == nil {
		return i
	}

	return def
}

func (j *JSONOBJ) MustUint32(args ...uint32) uint32 {
	var def uint32

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		panic(fmt.Sprintf("MustUint32() received too many arguments %d", len(args)))
	}

	i, err := j.Uint32()
	if err == nil {
		return i
	}
	return def
}

// Float64 coerces into a float64
func (j *JSONOBJ) Float64() (float64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(j.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int64 coerces into an int64
func (j *JSONOBJ) Int64() (int64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(j.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Uint64 coerces into an uint64
func (j *JSONOBJ) Uint64() (uint64, error) {
	switch j.data.(type) {
	case json.Number:
		return strconv.ParseUint(j.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(j.data).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}

// uint32 coerces into an uint32
func (j *JSONOBJ) Uint32() (uint32, error) {
	switch j.data.(type) {
	case json.Number:
		val, err := strconv.ParseUint(j.data.(json.Number).String(), 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(val), nil
	case float32, float64:
		return uint32(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint32(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return uint32(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

//-------------------------------------------------------------------------
func _isStringInArray(item string, arr []string) bool {
	for _, one := range arr {
		if item == one {
			return true
		}
	}

	return false
}

//验证key必须在范围内
func (j *JSONOBJ) VerifyMapKeyInArray(keys []string) error {
	m, err := j.Map()
	if err == nil {
		for k, _ := range m {
			if !_isStringInArray(k, keys) {
				return errors.New(k + " is not valid key")
			}
		}

		return nil
	}

	return errors.New("json object is not map")
}

//验证key必须存在
func (j *JSONOBJ) VerifyMapKeyExist(key ...string) error {
	m, err := j.Map()
	if err == nil {
		for _, k := range key {
			if _, ok := m[k]; !ok {
				return errors.New(k + " not exist")
			}
		}

		return nil
	}

	return errors.New("json object is not map")
}

//如果key存在，检查它是不是string类型
func (j *JSONOBJ) VerifyStringValue(key ...string) error {
	var err error

	for _, k := range key {
		if !j.IsKeyExist(k) {
			continue
		}

		_, err = j.Get(k).String()
		if err != nil {
			return errors.New(k + " is not string")
		}
	}

	return nil
}

//如果key存在，检查它是不是int64类型
func (j *JSONOBJ) VerifyInt64Value(key ...string) error {
	var err error

	for _, k := range key {
		if !j.IsKeyExist(k) {
			continue
		}

		_, err = j.Get(k).Int64()
		if err != nil {
			return errors.New(k + " is not int")
		}
	}

	return nil
}

//如果key存在，检查它是不是[]interface{}类型.如果key不存在则检查当前对象
func (j *JSONOBJ) VerifyArrayValue(key ...string) error {
	var err error

	if len(key) > 0 {
		for _, k := range key {
			if !j.IsKeyExist(k) {
				continue
			}

			_, err = j.Get(k).Array()
			if err != nil {
				return errors.New(k + " is not array")
			}
		}
	} else {
		if _, err = j.Array(); err != nil {
			return errors.New("not array")
		}
	}

	return nil
}

//如果key存在，检查它是不是[]map[string]interface{}类型.如果key不存在则检查当前对象
func (j *JSONOBJ) VerifyMapArrayValue(key ...string) error {
	var err error

	if len(key) > 0 {
		for _, k := range key {
			if !j.IsKeyExist(k) {
				continue
			}

			_, err = j.Get(k).MapArray()
			if err != nil {
				return errors.New(k + " is not valid array")
			}
		}
	} else {
		if _, err = j.MapArray(); err != nil {
			return errors.New("not valid array")
		}
	}

	return nil
}

//如果key存在，检查它是不是[]string类型.如果key不存在则检查当前对象
func (j *JSONOBJ) VerifyStringArrayValue(key ...string) error {
	var err error

	if len(key) > 0 {
		for _, k := range key {
			if !j.IsKeyExist(k) {
				continue
			}

			_, err = j.Get(k).StringArray()
			if err != nil {
				return errors.New(k + " is not string array")
			}
		}
	} else {
		if _, err = j.StringArray(); err != nil {
			return errors.New("not string array")
		}
	}

	return nil
}

//如果key存在，检查它是不是[]int64类型.如果key不存在则检查当前对象
func (j *JSONOBJ) VerifyInt64ArrayValue(key ...string) error {
	var err error

	if len(key) > 0 {
		for _, k := range key {
			if !j.IsKeyExist(k) {
				continue
			}

			_, err = j.Get(k).Int64Array()
			if err != nil {
				return errors.New(k + " is not string array")
			}
		}
	} else {
		if _, err = j.Int64Array(); err != nil {
			return errors.New("not int array")
		}
	}

	return nil
}

//如果key存在，则检查它是不是string类型, 且长度在范围[minLen, maxLen]内
func (j *JSONOBJ) VerifyStringLength(key string, minLen int, maxLen int) error {
	var err error

	if !j.IsKeyExist(key) {
		return nil
	}

	s, err := j.Get(key).String()
	if err != nil {
		return errors.New(key + " is not string")
	}

	nLen := len(s)
	if nLen < minLen || nLen > maxLen {
		return errors.New(key + " length error")
	}

	return nil
}

//如果key存在，检查它是不是int64类型, 且在范围[min, max]内
func (j *JSONOBJ) VerifyInt64Range(key string, min int64, max int64) error {
	var err error

	if !j.IsKeyExist(key) {
		return nil
	}

	nNum, err := j.Get(key).Int64()
	if err != nil {
		return errors.New(key + " is not int")
	}

	if nNum >= min && nNum <= max {
		return nil
	}

	return errors.New(key + " invalid")
}

//如果key存在，则检查它是不是string类型, 是否符合email名称要求
func (j *JSONOBJ) VerifyEmail(key string) error {
	var err error

	if !j.IsKeyExist(key) {
		return nil
	}

	s, err := j.Get(key).String()
	if err != nil {
		return errors.New(key + " is not valid email")
	}

	matched, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", s)
	if false == matched {
		return errors.New(key + " is not valid email")
	}

	return nil
}
