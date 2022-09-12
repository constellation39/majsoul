package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

// Hash password with hmac sha256
// return hash string
func Hash(data string) string {
	hash := hmac.New(sha256.New, []byte("lailai"))
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

// Save LoadFile config from file
func Save(path string, in interface{}) error {
	return SaveStruct(path, reflect.ValueOf(in))
}

func SaveStruct(path string, vTarget reflect.Value) error {
	oTarget := vTarget.Type()
	if oTarget.Elem().Kind() != reflect.Struct {
		return errors.New("type of received parameter is not struct")
	}
	data, err := json.Marshal(vTarget.Interface())
	if err != nil {
		return err
	}
	return SaveFile(path, data)
}

// SaveFile SaveFile config to file
func SaveFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, 0644)
}

func ExitsFile(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func Read(path string, data interface{}) error {
	if ExitsFile(path) {
		return fmt.Errorf("open %s error: File does not exist", path)
	}
	return LoadStruct(reflect.ValueOf(data), path)
}

func LoadStruct(vTarge reflect.Value, path string) error {
	oTarge := vTarge.Type()
	if oTarge.Elem().Kind() != reflect.Struct {
		return errors.New("type of received parameter is not struct")
	}
	data, err := LoadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, vTarge.Interface())
	if err != nil {
		return err
	}
	return nil
}

func LoadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
