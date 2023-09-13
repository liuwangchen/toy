package yaml

import (
	"github.com/liuwangchen/toy/transport/encoding"
	"gopkg.in/yaml.v2"
)

// Name is the name registered for the yaml codec.
const Name = "yaml"

func init() {
	encoding.RegisterCodec(codec{})
}

// codec is a Codec implementation with yaml.
type codec struct{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	return yaml.Marshal(v)
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func (codec) Name() string {
	return Name
}
