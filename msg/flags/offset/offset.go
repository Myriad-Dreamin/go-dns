package DNSFlags

type Type = uint16

const (
	QR     Type = 15
	OpCode Type = 11
	AA     Type = 10
	TC     Type = 9
	RD     Type = 8
	RA     Type = 7
	Z      Type = 4
	RCode  Type = 0
)
