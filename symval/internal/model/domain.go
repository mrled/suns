package model

import "time"

// SymmetryType represents the type of symmetry validation
type SymmetryType string

const (
	Palindrome      SymmetryType = "a"
	Flip180         SymmetryType = "b"
	DoubleFlip180   SymmetryType = "c"
	MirrorText      SymmetryType = "d"
	MirrorNames     SymmetryType = "e"
	AntonymNames    SymmetryType = "f"
)

// DomainData represents domain validation information
type DomainData struct {
	Owner        string
	Type         SymmetryType
	Hostname     string
	GroupID      string
	ValidateTime time.Time
}
