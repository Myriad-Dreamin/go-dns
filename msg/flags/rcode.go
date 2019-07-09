package DNSFlags

import offset "github.com/Myriad-Dreamin/go-dns/msg/flags/offset"

const (
	NoErrorCode Type = iota

	// FormatError:
	// 1) 发送的询问超过一个. 许多dns是这么做的，或者不回复信息。
	// 2) 发送的询问不能被解析，应当定期维护查看。一般来说发送者也不会发送错误的信息，
	//    可能主要是UDP

	FormatErrorCode

	// ServerFailure:
	// 1) 内部错误

	ServerFailureCode

	// NameError:
	// 1) 权威服务器特有的，因为我们的dns不管辖，所以直接转发。
	// 2) 又称 NXDomain

	NameErrorCode

	// NotImplemented:
	// 1) 直接回复。
	// 1) 直接转发。

	NotImplementedCode

	// Refused:
	// 1) 无需理由。
	// 1) 前面的一些特定错误也可以此回复。

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
