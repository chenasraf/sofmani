// Code generated from Pkl module `dev.casraf.sofmani.AppConfig`. DO NOT EDIT.
package platform

import (
	"encoding"
	"fmt"
)

type Platform string

const (
	Macos   Platform = "macos"
	Linux   Platform = "linux"
	Windows Platform = "windows"
)

// String returns the string representation of Platform
func (rcv Platform) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(Platform)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for Platform.
func (rcv *Platform) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "macos":
		*rcv = Macos
	case "linux":
		*rcv = Linux
	case "windows":
		*rcv = Windows
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid Platform`, str)
	}
	return nil
}
