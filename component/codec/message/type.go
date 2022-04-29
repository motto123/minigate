package message

import "fmt"

type Type byte

func (t Type) String() string {
	v := types[t]
	if v == "" {
		return fmt.Sprintf("Unknown %d", t)
	}
	return v
}

type DataType byte

func (t DataType) String() string {
	v := dataTypes[t]
	if v == "" {
		return fmt.Sprintf("Unknown %d", t)
	}
	return v
}
