//go:build sugar

package contestio

// Нужно только, если есть желание пользоваться обобщенными функциями ScanSlice/PrintSlice (см. sugar.go)
type Types []Type

func (s Types) Slice() []Type                     { return s }
func (s Types) Parse(b []byte) (Type, error)      { return parseType(b) }
func (s Types) AppendVal(b []byte, v Type) []byte { return append(b, v...) }
