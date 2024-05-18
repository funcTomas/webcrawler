package reader

import (
	"bytes"
	"fmt"
	"io"
)

type MultipleReader interface {
	Reader() io.ReadCloser
}

type myMultipleReader struct {
	data []byte
}

func (mReader *myMultipleReader) Reader() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(mReader.data))
}
func NewMultipleReader(reader io.Reader) (MultipleReader, error) {
	var data []byte
	var err error
	if reader != nil {
		data, err = io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("multiple reader: could not create a new one: %s", err)
		}
	} else {
		data = []byte{}
	}
	return &myMultipleReader{
		data: data,
	}, nil
}
