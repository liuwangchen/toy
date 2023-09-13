package config

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"path"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/a8m/envsubst"
	"gopkg.in/yaml.v2"
)

func unmarshal(ext string, buf []byte, v interface{}) error {
	switch ext {
	case ".xml":
		if err := xml.Unmarshal(buf, v); err != nil {
			return err
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(buf, v); err != nil {
			return err
		}
	case ".toml":
		if err := toml.Unmarshal(buf, v); err != nil {
			return err
		}
	case ".json":
		if err := json.Unmarshal(buf, v); err != nil {
			return err
		}
	default:
		if err := json.Unmarshal(buf, v); err != nil {
			return err
		}
	}
	return nil
}

type IConfigSource interface {
	Load(v interface{}, contentsHandle func([]byte) []byte) error
}

// FileConfigSource 数据源是本地文件
type FileConfigSource struct {
	FilePath string // 文件路径
}

func (f *FileConfigSource) Load(v interface{}, contentsHandle func([]byte) []byte) error {
	contents, err := ioutil.ReadFile(f.FilePath)
	if err != nil {
		return err
	}

	ext := path.Ext(f.FilePath)
	envBuf, err := envsubst.Bytes(contents)
	if err != nil {
		return err
	}
	envBuf = contentsHandle(envBuf)
	var m = new(map[string]interface{})
	err = unmarshal(ext, envBuf, m)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	tmpl, err := template.New("").Parse(string(envBuf))
	if err != nil {
		return err
	}
	err = tmpl.Execute(&b, m)
	if err != nil {
		return err
	}
	return unmarshal(ext, b.Bytes(), v)
}

// loadConfigFromSource 加载配置
func loadConfigFromSource(ics IConfigSource, configValue interface{}, contentsHandle func([]byte) []byte) error {
	return ics.Load(configValue, contentsHandle)
}

// LoadConfigFromFile 从本地文件加载配置
func LoadConfigFromFile(filePath string, configValue interface{}, contentsHandle func([]byte) []byte) error {
	return loadConfigFromSource(&FileConfigSource{
		FilePath: filePath,
	}, configValue, contentsHandle)
}
