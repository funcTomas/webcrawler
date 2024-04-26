package field

type FieldType int

const (
	UnknownType FieldType = 0
	BoolType    FieldType = 1
	Int64Type   FieldType = 2
	Float64Type FieldType = 3
	StringType  FieldType = 4
	ObjectType  FieldType = 5
)

type Field interface {
	Name() string
	Type() FieldType
	Value() interface{}
}
