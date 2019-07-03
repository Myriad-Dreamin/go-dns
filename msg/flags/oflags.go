package DNSFlags

import offset "github.com/Myriad-Dreamin/go-dns/msg/flags/offset"

type Type = uint16

const (
	AA uint16 = 1 << offset.AA
	TC uint16 = 1 << offset.TC
	RD uint16 = 1 << offset.RD
	RA uint16 = 1 << offset.RA
)

func HasAA(flags uint16) bool {
	return (flags & AA) != 0
}

func HasTC(flags uint16) bool {
	return (flags & TC) != 0
}

func HasRD(flags uint16) bool {
	return (flags & RD) != 0
}

func HasRA(flags uint16) bool {
	return (flags & RA) != 0
}
