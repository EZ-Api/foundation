package jsoncodec

import (
	"io"

	"github.com/bytedance/sonic"
)

func Marshal(v any) ([]byte, error) { return sonic.Marshal(v) }

func Unmarshal(data []byte, v any) error { return sonic.Unmarshal(data, v) }

func UnmarshalString(data string, v any) error { return sonic.UnmarshalString(data, v) }

func NewEncoder(w io.Writer) sonic.Encoder { return sonic.ConfigDefault.NewEncoder(w) }

func NewDecoder(r io.Reader) sonic.Decoder { return sonic.ConfigDefault.NewDecoder(r) }

