package model

import "time"

// SymmetryType represents the type of symmetry validation
type SymmetryType string

const (
	Palindrome      SymmetryType = "palindrome"
	Flip180         SymmetryType = "180flip"
	DoubleFlip180   SymmetryType = "double180flip"
	MirrorText      SymmetryType = "mirrortext"
	MirrorNames     SymmetryType = "mirrornames"
	AntonymNames    SymmetryType = "antonymnames"
)

// DomainData represents domain validation information
type DomainData struct {
	ValidateTime time.Time
	Owner        string
	Domain       string
	Flip         *string
	Type         SymmetryType
}
