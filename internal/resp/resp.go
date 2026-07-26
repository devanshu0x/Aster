package resp

type RESPType string

const(
	RESPSimpleString RESPType = "simple_string"
	RESPBulkString RESPType="bulk_string"
	RESPError RESPType="error"
	RESPInteger RESPType="integer"
	RESPArray RESPType= "array"
)

type RESPValue struct{
	Type RESPType
	Value interface{}
}