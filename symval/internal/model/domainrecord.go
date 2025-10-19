package model

import (
	"time"

	"github.com/mrled/suns/symval/internal/symgroup"
)

// DomainRecord represents domain validation information
type DomainRecord struct {
	Owner        string
	Type         symgroup.SymmetryType
	Hostname     string
	GroupID      string
	ValidateTime time.Time
}
