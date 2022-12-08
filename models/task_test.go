package models

import (
	"fmt"
	"testing"

	. "github.com/stretchr/testify/require"
)

type Data struct {
	A string
	b *string
}

func TestStatus(t *testing.T) {
	var (
		bts    []byte
		err    error
		status Status
	)

	status = Created
	bts, err = status.MarshalJSON()
	NoError(t, err)
	EqualValues(t, bts, []byte("created"))

	status = Status("unknown")
	_, err = status.MarshalJSON()
	Error(t, err)
}

func TestData(t *testing.T) {
	s1 := &Data{A: "hello", b: new(string)}

	s2 := *s1
	fmt.Printf("s2.b is nil: %t\n", s2.b == nil) // false
}
