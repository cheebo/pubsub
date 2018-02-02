package gob_test

import (
	"github.com/cheebo/pubsub/marshalling/gob"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Message struct {
	Name   string
	Age    int
	Salary float64
}

func TestMarshallUnMarshall(t *testing.T) {
	asserts := assert.New(t)

	data := Message{
		Name:   "John Doe",
		Age:    25,
		Salary: 10000.37,
	}

	bytes, err := gob.Marshall(data)
	asserts.NoError(err)

	var msg Message
	gob.UnMarshall(bytes, &msg)

	asserts.EqualValues(data, msg)
}
