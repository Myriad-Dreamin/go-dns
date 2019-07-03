package DNSFlags

import offset "github.com/Myriad-Dreamin/go-dns/msg/flags/offset"

const (
	QCode Type = iota
	RCode
	QROffset uint16 = 1 << offset.QR
)

const (
	Q Type = iota << offset.QR
	R
)

func HasQ(flags uint16) bool {
	return ((flags >> offset.QR) & 0x1) == QCode
}

func HasR(flags uint16) bool {
	return ((flags >> offset.QR) & 0x1) == RCode
}

func SetR(flags *uint16) {
	*flags |= 0x1 << offset.QR
}

func SetQ(flags *uint16) {
	*flags &^= 0x1 << offset.QR
}
