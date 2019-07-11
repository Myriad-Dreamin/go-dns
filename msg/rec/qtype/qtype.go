package QType

type Type = uint16

// <domain-name> is a domain name represented as a series of labels, and
// terminated by a label with zero length.  <character-string> is a single
// length octet followed by that number of characters.  <character-string>
// is treated as binary information, and can be up to 256 characters in
// length (including the length octet).

const (

	// 1 0x1
	A Type = iota + 1

	// 3.3.11. NS RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   NSDNAME                     /
	//     /                <domain-name>                  /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// NSDNAME         A <domain-name> which specifies a host which should be
	//                 authoritative for the specified class and domain.
	//
	// NS records cause both the usual additional section processing to locate
	// a type A record, and, when used in a referral, a special search of the
	// zone in which they reside for glue information.
	//
	// The NS RR states that the named host should be expected to have a zone
	// starting at owner name of the specified class.  Note that the class may
	// not indicate the protocol family which should be used to communicate
	// with the host, although it is typically a strong hint.  For example,
	// hosts which are name servers for either Internet (IN) or Hesiod (HS)
	// class information are normally queried using IN class protocols.
	// 2 0x2
	NS

	// Obsolete
	// 3 0x3
	MD

	// Obsolete
	// 4 0x4
	MF

	//
	// 3.3.1. CNAME RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                     CNAME                     /
	//     /                 <domain-name>                 /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// CNAME           A <domain-name> which specifies the canonical or primary
	//                 name for the owner.  The owner name is an alias.
	//
	// CNAME RRs cause no additional section processing, but name servers may
	// choose to restart the query at the canonical name in certain cases.  See
	// the description of name server logic in [RFC-1034] for details.
	// 5 0x5
	CNAME

	// 3.3.13. SOA RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                     MNAME                     /
	//     /                 <domain-name>                 /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                     RNAME                     /
	//     /                 <domain-name>                 /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                    SERIAL                     |
	//     |                <4-bytes int>                  |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                    REFRESH                    |
	//     |                <4-bytes int>                  |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                     RETRY                     |
	//     |                <4-bytes int>                  |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                    EXPIRE                     |
	//     |                <4-bytes int>                  |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                    MINIMUM                    |
	//     |                <4-bytes int>                  |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// MNAME           The <domain-name> of the name server that was the
	//                 original or primary source of data for this zone.
	//
	// RNAME           A <domain-name> which specifies the mailbox of the
	//                 person responsible for this zone.
	//
	// SERIAL          The unsigned 32 bit version number of the original copy
	//                 of the zone.  Zone transfers preserve this value.  This
	//                 value wraps and should be compared using sequence space
	//                 arithmetic.
	//
	// REFRESH         A 32 bit time interval before the zone should be
	//                 refreshed.
	//
	// RETRY           A 32 bit time interval that should elapse before a
	//                 failed refresh should be retried.
	//
	// EXPIRE          A 32 bit time value that specifies the upper limit on
	//                 the time interval that can elapse before the zone is no
	//                 longer authoritative.
	// MINIMUM         The unsigned 32 bit minimum TTL field that should be
	//                 exported with any RR from this zone.
	//
	// SOA records cause no additional section processing.
	//
	// All times are in units of seconds.
	//
	// Most of these fields are pertinent only for name server maintenance
	// operations.  However, MINIMUM is used in all query operations that
	// retrieve RRs from a zone.  Whenever a RR is sent in a response to a
	// query, the TTL field is set to the maximum of the TTL field from the RR
	// and the MINIMUM field in the appropriate SOA.  Thus MINIMUM is a lower
	// bound on the TTL field for all RRs in a zone.  Note that this use of
	// MINIMUM should occur when the RRs are copied into the response and not
	// when the zone is loaded from a master file or via a zone transfer.  The
	// reason for this provison is to allow future dynamic update facilities to
	// change the SOA RR with known semantics.
	// 6 0x6
	SOA

	// 3.3.3. MB RDATA format (EXPERIMENTAL)
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   MADNAME                     /
	//     /                                               /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// MADNAME         A <domain-name> which specifies a host which has the
	//                 specified mailbox.
	//
	// MB records cause additional section processing which looks up an A type
	// RRs corresponding to MADNAME.
	// 7 0x7
	MB

	// 3.3.6. MG RDATA format (EXPERIMENTAL)
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   MGMNAME                     /
	//     /                                               /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// MGMNAME         A <domain-name> which specifies a mailbox which is a
	//                 member of the mail group specified by the domain name.
	//
	// MG records cause no additional section processing.
	// 8 0x8
	MG

	// 3.3.8. MR RDATA format (EXPERIMENTAL)
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   NEWNAME                     /
	//     /                                               /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// NEWNAME         A <domain-name> which specifies a mailbox which is the
	//                 proper rename of the specified mailbox.
	//
	// MR records cause no additional section processing.  The main use for MR
	// is as a forwarding entry for a user who has moved to a different
	// mailbox.
	// 9 0x9
	MR

	// (Experimental)
	// NULL records cause no additional section processing.  NULL RRs are not
	// allowed in master files.  NULLs are used as placeholders in some
	// experimental extensions of the DNS.
	// 10 0xa
	NULL

	// 	3.4.2. WKS RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                    ADDRESS                    |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |       PROTOCOL        |                       |
	//     +--+--+--+--+--+--+--+--+                       |
	//     |                                               |
	//     /                   <BIT MAP>                   /
	//     /                                               /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// ADDRESS         An 32 bit Internet address
	//
	// PROTOCOL        An 8 bit IP protocol number
	//
	// <BIT MAP>       A variable length bit map.  The bit map must be a
	//                 multiple of 8 bits long.
	//
	// The WKS record is used to describe the well known services supported by
	// a particular protocol on a particular internet address.  The PROTOCOL
	// field specifies an IP protocol number, and the bit map has one bit per
	// port of the specified protocol.  The first bit corresponds to port 0,
	// the second to port 1, etc.  If the bit map does not include a bit for a
	// protocol of interest, that bit is assumed zero.  The appropriate values
	// and mnemonics for ports and protocols are specified in [RFC-1010].
	//
	// For example, if PROTOCOL=TCP (6), the 26th bit corresponds to TCP port
	// 25 (SMTP).  If this bit is set, a SMTP server should be listening on TCP
	// port 25; if zero, SMTP service is not supported on the specified
	// address.
	//
	// The purpose of WKS RRs is to provide availability information for
	// servers for TCP and UDP.  If a server supports both TCP and UDP, or has
	// multiple Internet addresses, then multiple WKS RRs are used.
	//
	// WKS RRs cause no additional section processing.
	//
	// In master files, both ports and protocols are expressed using mnemonics
	// or decimal numbers.
	// 11 0xb
	WKS

	// 3.3.12. PTR RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   PTRDNAME                    /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// PTRDNAME        A <domain-name> which points to some location in the
	//                 domain name space.
	//
	// PTR records cause no additional section processing.  These RRs are used
	// in special domains to point to some other location in the domain space.
	// These records are simple data, and don't imply any special processing
	// similar to that performed by CNAME, which identifies aliases.  See the
	// description of the IN-ADDR.ARPA domain for an example.
	// 12 0xc
	PTR

	// 3.3.2. HINFO RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /            CPU  <character-string>            /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /            OS   <character-string>            /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// CPU             A <character-string> which specifies the CPU type.
	//
	// OS              A <character-string> which specifies the operating
	//                 system type.
	//
	// Standard values for CPU and OS can be found in [RFC-1010].
	//
	// HINFO records are used to acquire general information about a host.  The
	// main use is for protocols such as FTP that can use special procedures
	// when talking between machines or operating systems of the same type.
	// 13 0xd
	HINFO

	// 3.3.7. MINFO RDATA format (EXPERIMENTAL)
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                    RMAILBX                    /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                    EMAILBX                    /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// RMAILBX         A <domain-name> which specifies a mailbox which is
	//                 responsible for the mailing list or mailbox.  If this
	//                 domain name names the root, the owner of the MINFO RR is
	//                 responsible for itself.  Note that many existing mailing
	//                 lists use a mailbox X-request for the RMAILBX field of
	//                 mailing list X, e.g., Msgroup-request for Msgroup.  This
	//                 field provides a more general mechanism.
	//
	//
	// EMAILBX         A <domain-name> which specifies a mailbox which is to
	//                 receive error messages related to the mailing list or
	//                 mailbox specified by the owner of the MINFO RR (similar
	//                 to the ERRORS-TO: field which has been proposed).  If
	//                 this domain name names the root, errors should be
	//                 returned to the sender of the message.
	//
	// MINFO records cause no additional section processing.  Although these
	// records can be associated with a simple mailbox, they are usually used
	// with a mailing list.
	// 14 0xe
	MINFO

	// 3.3.9. MX RDATA format

	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     |                  PREFERENCE                   |
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   EXCHANGE                    /
	//     /                                               /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// PREFERENCE      A 16 bit integer which specifies the preference given to
	//                 this RR among others at the same owner.  Lower values
	//                 are preferred.
	//
	// EXCHANGE        A <domain-name> which specifies a host willing to act as
	//                 a mail exchange for the owner name.
	//
	// MX records cause type A additional section processing for the host
	// specified by EXCHANGE.  The use of MX RRs is explained in detail in
	// [RFC-974].
	// 15 0xf
	MX

	// 3.3.14. TXT RDATA format
	//
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//     /                   TXT-DATA                    /
	//     +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	//
	// where:
	//
	// TXT-DATA        One or more <character-string>s.
	//
	// TXT RRs are used to hold descriptive text.  The semantics of the text
	// depends on the domain where it is found.
	// 16 0x10
	TXT

	AAAA Type = 28

	// 33 0x21
	SRV Type = 33

	// 38 0x26
	A6 Type = 38

	OPT Type = 41

	AXFR Type = iota + 252
	MALIB
	MALIA
	Asterisk
)
