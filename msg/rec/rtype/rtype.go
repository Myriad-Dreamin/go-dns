package RType

type Type = uint16

const (
	A Type = iota + 1 // 1
	NS
	MD
	MF
	CNAME
	SOA // 6
	MB
	MG
	MR
	// (Experimental)
	NULL
	WKS // b
	PTR
	HINFO
	MINFO
	MX
	TXT // 10

	AAAA Type = 28
	OPT  Type = 41
)
