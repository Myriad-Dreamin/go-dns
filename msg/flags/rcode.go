package DNSFlags

import offset "github.com/Myriad-Dreamin/go-dns/msg/flags/offset"

const (
	NoErrorCode Type = iota
	FormatErrorCode
	ServerFailureCode
	NameErrorCode
	NotImplementedCode
	RefusedCode
)

const (
	RCodeOffset uint16 = 1 << offset.RCode
	NoError     uint16 = iota << offset.RCode
	FormatError
	ServerFailure
	NameError
	NotImplemented
	Refused
)

func HasNoError(flags uint16) bool {
	return ((flags >> offset.RCode) & 0xf) == NoErrorCode
}

func SetNoError(flags *uint16) {
	*flags &^= (0xf << offset.RCode)
	*flags |= NoError
}

func HasFormatError(flags uint16) bool {
	return ((flags >> offset.RCode) & 0xf) == FormatErrorCode
}

func SetFormatError(flags *uint16) {
	*flags &^= (0xf << offset.RCode)
	*flags |= FormatError
}

func HasServerFailure(flags uint16) bool {
	return ((flags >> offset.RCode) & 0xf) == ServerFailureCode
}

func SetServerFailure(flags *uint16) {
	*flags &^= (0xf << offset.RCode)
	*flags |= ServerFailure
}

func HasNameError(flags uint16) bool {
	return ((flags >> offset.RCode) & 0xf) == NameErrorCode
}

func SetNameError(flags *uint16) {
	*flags &^= (0xf << offset.RCode)
	*flags |= NameError
}

func HasNotImplemented(flags uint16) bool {
	return ((flags >> offset.RCode) & 0xf) == NotImplementedCode
}

func SetNotImplemented(flags *uint16) {
	*flags &^= (0xf << offset.RCode)
	*flags |= NotImplemented
}

func HasRefused(flags uint16) bool {
	return ((flags >> offset.RCode) & 0xf) == RefusedCode
}

func SetRefused(flags *uint16) {
	*flags &^= (0xf << offset.RCode)
	*flags |= Refused
}
