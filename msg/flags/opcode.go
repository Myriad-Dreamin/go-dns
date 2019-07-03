package DNSFlags

import offset "github.com/Myriad-Dreamin/go-dns/msg/flags/offset"

const (
	QueryCode Type = iota
	IQueryCode
	StatusCode
)
const (
	OpCodeOffset uint16 = 1 << offset.OpCode
	Query        Type   = iota << offset.OpCode
	IQuery
	Status
)

func HasQuery(flags uint16) bool {
	return ((flags >> offset.OpCode) & 0xf) == QueryCode
}

func HasIQuery(flags uint16) bool {
	return ((flags >> offset.OpCode) & 0xf) == IQueryCode
}

func HasStatus(flags uint16) bool {
	return ((flags >> offset.OpCode) & 0xf) == StatusCode
}

func SetQuery(flags *uint16) {
	*flags &^= (0xf << offset.OpCode)
	*flags |= Query
}

func SetIQuery(flags *uint16) {
	*flags &^= (0xf << offset.OpCode)
	*flags |= IQuery
}

func SetStatus(flags *uint16) {
	*flags &^= (0xf << offset.OpCode)
	*flags |= Status
}

func UnsetOpCode(flags *uint16) {
	*flags &^= (0xf << offset.OpCode)
}
