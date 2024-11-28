// Code generated from Pkl module `dev.casraf.sofmani.AppConfig`. DO NOT EDIT.
package appconfig

import "github.com/chenasraf/sofmani/appconfig/installertype"

type Installer struct {
	Name string `pkl:"name"`

	Type installertype.InstallerType `pkl:"type"`

	Platforms *Platforms `pkl:"platforms"`

	Url *string `pkl:"url"`

	Command *string `pkl:"command"`

	Steps *[]*Installer `pkl:"steps"`
}
