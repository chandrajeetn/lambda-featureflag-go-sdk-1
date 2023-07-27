package logger

import (
	"github.com/sirupsen/logrus"
)

type ContextFields logrus.Fields


func (ctxFields ContextFields) Set(key string, data interface{}) {
	ctxFields[key] = data
}

func (ctxFields ContextFields) Append(data logrus.Fields) {
	for k, v := range data {
		vm, ok := v.(map[string]interface{})
		if ok {
			ctxFields[k] = DeepCopyMap(vm)
		} else {
			ctxFields[k] = v
		}
	}
}

//func (c ContextFields) Get(key string)  interface{} {
//	return c[key]
//}
//
//func (c ContextFields) Delete(key string) bool{
//	delete(c, key)
//	return true
//}

func (ctxFields ContextFields) GetAll()  logrus.Fields {
	logrusFields := logrus.Fields{}
	contextFieldsData:=DeepCopyMap(ctxFields)
	for k, v := range contextFieldsData {
		vm, ok := v.(map[string]interface{})
		if ok {
			logrusFields[k] = DeepCopyMap(vm)
		} else {
			logrusFields[k] = v
		}
	}
	return logrusFields

}


func DeepCopyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = DeepCopyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}
