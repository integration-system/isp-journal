package entry

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	a := assert.New(t)

	data := bytes.NewBuffer(make([]byte, 0, 4096))
	bytes, err := MarshalToBytes(e)
	if err != nil {
		panic(err)
	}
	data.Write(bytes)

	copy := *e
	copy.Response = []byte("")
	copy.Request = []byte("111")
	bytes, err = MarshalToBytes(&copy)
	if err != nil {
		panic(err)
	}
	data.Write(bytes)

	b0, _ := json.Marshal(e)
	copy0, _ := json.Marshal(copy)

	e1, err := UnmarshalNext(data)
	if err != nil {
		panic(err)
	}
	b1, _ := json.Marshal(e1)
	a.EqualValues(b0, b1)

	e2, err := UnmarshalNext(data)
	if err != nil {
		panic(err)
	}
	b2, _ := json.Marshal(e2)
	a.EqualValues(copy0, b2)
}

func BenchmarkUnmarshalNext(b *testing.B) {
	bs, err := MarshalToBytes(e)
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(bs)
		_, err := UnmarshalNext(buffer)
		if err != nil {
			panic(err)
		}
	}
}
