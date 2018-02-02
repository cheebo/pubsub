package gob

import (
	"bytes"
	"encoding/gob"
)

func Marshall(data interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)

	err := enc.Encode(data)
	return b.Bytes(), err
}

/*
 * @param result - pointer
 */
func UnMarshall(data []byte, result interface{}) error {
	b := new(bytes.Buffer)
	b.Write(data)
	d := gob.NewDecoder(b)
	return d.Decode(result)
}
