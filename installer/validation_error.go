package installer

import "fmt"

// ValidationError represents a validation error for an installer configuration.
type ValidationError struct {
	// FieldName is the name of the field that failed validation.
	FieldName string
	// Message is a description of the validation error.
	Message string
	// InstallerName is the name of the installer where the validation error occurred.
	InstallerName string
}

// Error returns a string representation of the validation error.
func (v ValidationError) Error() string {
	return fmt.Sprintf("Validation Error in %s - Field '%s' is invalid: %s.", v.InstallerName, v.FieldName, v.Message)
}

// validationIsRequired returns a standard message for a required field.
func validationIsRequired() string {
	return "Must be specified"
}

// validationInvalidFormat returns a standard message for an invalid format.
func validationInvalidFormat() string {
	return "Invalid format"
}

// validationIsNotEmpty returns a standard message for a field that cannot be empty.
func validationIsNotEmpty() string {
	return "Cannot be empty"
}
