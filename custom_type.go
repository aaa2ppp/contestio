package contestio

// Пример создания кастомного типа.

type Type  string

// Вам нужно реализовать свои функции parseType и appendType.
//
// ВАЖНО: token это сырые данные из буфера Reader, действительные только до следующего чтения.
// Вы всегда должы копировать их при необходимости сохранить.
// Здесь Type(token) эквивалентно string(slice_of_bytes), что в Go всегда приводи к копированию.

func parseType(token []byte) (Type, error) { return Type(token), nil }
func appendType(b []byte, v Type) []byte   { return append(b, v...) }

// Остальной код шаблонный (просто замените подстроку "Type" на имя вашего типа).

func ScanTypes(br *Reader, a []Type) (int, error)      { return scanSlice(br, parseType, a) }
func ScanTypesLn(br *Reader, a []Type) ([]Type, error) { return scanSliceLn(br, parseType, a) }
func ScanType(br *Reader, a ...*Type) (int, error)     { return scanVars(br, parseType, a...) }
func ScanTypeLn(br *Reader, a ...*Type) (int, error)   { return scanVarsLn(br, parseType, a...) }

func PrintTypes(bw *Writer, op WO, a []Type) (int, error) { return printSlice(bw, op, appendType, a) }
func PrintTypesLn(bw *Writer, a []Type) (int, error)      { return printSliceLn(bw, appendType, a) }
func PrintType(bw *Writer, op WO, a ...Type) (int, error) { return printVals(bw, op, appendType, a...) }
func PrintTypeLn(bw *Writer, a ...Type) (int, error)      { return printValsLn(bw, appendType, a...) }

// Нужно только, если есть желание пользоваться обобщенными функциями ScanSlice/PrintSlice (см. sugar.go)
type Types []Type
func (s Types) Slice() []Type                     { return s }
func (s Types) Parse(b []byte) (Type, error)      { return parseType(b) }
func (s Types) AppendVal(b []byte, v Type) []byte { return append(b, v...) }

