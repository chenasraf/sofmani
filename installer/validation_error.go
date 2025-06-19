package installer

import "fmt"

type ValidationError struct {
	FieldName     string
	Message       string
	InstallerName string
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("Validation Error in %s - Field '%s' is invalid: %s.", v.InstallerName, v.FieldName, v.Message)
}

func validationIsRequired() string {
	return "Must be specified"
}

func validationInvalidFormat() string {
	return "Invalid format"
}

func validationIsNotEmpty() string {
	return "Cannot be empty"
}
