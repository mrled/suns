package symgroup

// SymmetryType represents the type of symmetry validation
type SymmetryType string

const (
	Palindrome    SymmetryType = "a"
	Flip180       SymmetryType = "b"
	DoubleFlip180 SymmetryType = "c"
	MirrorText    SymmetryType = "d"
	MirrorNames   SymmetryType = "e"
	AntonymNames  SymmetryType = "f"
)

// TypeNameToCode maps human-readable type names to their single-character codes
var TypeNameToCode = map[string]string{
	"palindrome":    "a",
	"flip180":       "b",
	"doubleflip180": "c",
	"mirrortext":    "d",
	"mirrornames":   "e",
	"antonymnames":  "f",
}

// TypeCodeToName maps single-character codes to their human-readable names
var TypeCodeToName = map[string]string{
	"a": "palindrome",
	"b": "flip180",
	"c": "doubleflip180",
	"d": "mirrortext",
	"e": "mirrornames",
	"f": "antonymnames",
}
