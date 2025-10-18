package model

import (
	"time"

	"github.com/mrled/suns/symval/internal/symgroup"
)

// DomainData represents domain validation information
type DomainData struct {
	Owner        string
	Type         symgroup.SymmetryType
	Hostname     string
	GroupID      string
	ValidateTime time.Time
}
