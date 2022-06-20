package majsoul

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

// LoadConfig loadFile config from file
func LoadConfig(path string) (*Config, error) {
	cfg := new(Config)
	err := read(path, cfg)
	return cfg, err
}

func exitsFile(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func read(path string, data interface{}) error {
	if exitsFile(path) {
		return fmt.Errorf("open %s error: File does not exist", path)
	}
	return loadFile(reflect.ValueOf(data), path)
}

func loadFile(vTarge reflect.Value, path string) error {
	oTarge := vTarge.Type()
	if oTarge.Elem().Kind() != reflect.Struct {
		return errors.New("type of received parameter is not struct")
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, vTarge.Interface())
	if err != nil {
		return err
	}
	return nil
}
