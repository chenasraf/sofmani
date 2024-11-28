// Code generated from Pkl module `dev.casraf.sofmani.AppConfig`. DO NOT EDIT.
package appconfig

import "github.com/chenasraf/sofmani/appconfig/platform"

type Platforms struct {
	Only []platform.Platform `pkl:"only"`

	Except []platform.Platform `pkl:"except"`
}
