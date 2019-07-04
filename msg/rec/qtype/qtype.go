package QType

type Type = uint16

const (
	A Type = iota + 1
	NS
	MD
	MF
	CNAME
	SOA
	MB
	MG
	MR
	// (Experimental)
	NULL
	WKS
	PTR
	HINFO
	MINFO
	MX
	TXT

	AXFR Type = iota + 252
	MALIB
	MALIA
	Asterisk
)
