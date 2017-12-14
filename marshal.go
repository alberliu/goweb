package goweb

import "json-iterator/go"

type API interface {
	Unmarshal([]byte, interface{}) error
	Marshal(interface{}) ([]byte, error)
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func jsonMarshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}
