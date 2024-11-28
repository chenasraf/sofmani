// Code generated from Pkl module `dev.casraf.sofmani.AppConfig`. DO NOT EDIT.
package installertype

import (
	"encoding"
	"fmt"
)

type InstallerType string

const (
	Custom InstallerType = "custom"
	Brew   InstallerType = "brew"
	Apt    InstallerType = "apt"
	Git    InstallerType = "git"
	Unzip  InstallerType = "unzip"
)

// String returns the string representation of InstallerType
func (rcv InstallerType) String() string {
	return string(rcv)
}

var _ encoding.BinaryUnmarshaler = new(InstallerType)

// UnmarshalBinary implements encoding.BinaryUnmarshaler for InstallerType.
func (rcv *InstallerType) UnmarshalBinary(data []byte) error {
	switch str := string(data); str {
	case "custom":
		*rcv = Custom
	case "brew":
		*rcv = Brew
	case "apt":
		*rcv = Apt
	case "git":
		*rcv = Git
	case "unzip":
		*rcv = Unzip
	default:
		return fmt.Errorf(`illegal: "%s" is not a valid InstallerType`, str)
	}
	return nil
}
