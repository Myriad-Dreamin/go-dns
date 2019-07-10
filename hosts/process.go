package hosts

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"

	config "github.com/Myriad-Dreamin/go-dns/config"
)

var (
	errTooMuchItem = errors.New("too much item in a line")
	errIPFormat    = errors.New("bad format of ip")
	errFQDNFormat  = errors.New("bad format of domain name")
)

var (
	HostsIPv4 map[string]net.IP
	HostsIPv6 map[string]net.IP
)

type HostMapping = map[string]net.IP

func findChar(l []byte, c byte) int {
	for i := 0; i < len(l); i++ {
		if l[i] == c {
			return i
		}
	}
	return len(l)
}

const (
	// test62NRegExp = "|"
	// test63NRegExp = "^([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z])))$"

	test0to255 = "([0-9]|([1-9][0-9])|(1([0-9]{2}))|(2[0-4][0-9])|(25[0-5]))"
)

/*
<domain> ::= <subdomain> | " "

<subdomain> ::= <label> | <subdomain> "." <label>

<label> ::= <letter> [ [ <ldh-str> ] <let-dig> ]

<ldh-str> ::= <let-dig-hyp> | <let-dig-hyp> <ldh-str>

<let-dig-hyp> ::= <let-dig> | "-"

<let-dig> ::= <letter> | <digit>

<letter> ::= any one of the 52 alphabetic characters A through Z in
upper case and a through z in lower case

<digit> ::= any one of the ten digits 0 through 9

labels          63 octets or less
names           255 octets or less

in golang:

	label := [:alpha:] [-[:alnum:]]* [:alnum:]{0:1}
	subdomain := (<label> \.)* <label>
	domain := " " | <subdomain>



*/

var (
	testipv4 = "^(" + test0to255 + "\\." + test0to255 + "\\." + test0to255 + "\\." + test0to255 + ")$"
	regFQDN  = regexp.MustCompile("^(([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z])))|((([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z]))\\.)*)([0-9A-Za-z](()|([-0-9A-Za-z]{0,60}[0-9A-Za-z])))))$")
	regIP    = regexp.MustCompile(testipv4)
	// rega    = regexp.MustCompile(test63NRegExp)
	// regs    = regexp.MustCompile("^(([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z]))\\.)*)([0-9A-Za-z](()|([-0-9A-Za-z]{0,60}[0-9A-Za-z])))$")
)

func toIP(ipaddr []byte) net.IP {
	return net.ParseIP(string(ipaddr))
}

func isIP(ipaddr []byte) bool {
	return net.ParseIP(string(ipaddr)) != nil
}

func isIPv4(ipaddr []byte) bool {
	return regIP.Match(ipaddr)
}

func testFQDN(fqdn []byte) bool {
	if len(fqdn) > 255 {
		return false
	}
	return regFQDN.Match(fqdn)
}

// func testFQDNx(fqdn []byte) bool {
// 	return regs.Match(fqdn) || rega.Match(fqdn)
// }

// func ToIPv4Bytes(net.I)

// process any file to ipv4 & ipv6 map
func Process(filePath string) (HostMapping, HostMapping, error) {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	// var buf = bytes.NewBuffer(make([]byte, 0, 1024))
	var sp int
	var ipBytes, fqdnBytes, line []byte
	var ret4 = make(HostMapping)
	var ret6 = make(HostMapping)
	var mip net.IP
	for {
		line, err = rd.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if err == io.EOF {
			err = nil
			if len(line) == 0 {
				break
			}

			if line[0] == '#' {
				break
			}
			sp = findChar(line, ' ')
			ipBytes, line = line[0:sp], bytes.TrimSpace(line[sp:])
			if mip = toIP(ipBytes); mip == nil {
				return nil, nil, errIPFormat
			}
			if len(line) == 0 {
				err = errFQDNFormat
				break
			}

			for {
				if len(line) == 0 || line[0] == '#' {
					break
				}
				sp = findChar(line, ' ')
				fqdnBytes, line = line[0:sp], bytes.TrimSpace(line[sp:])

				if !testFQDN(fqdnBytes) {
					return nil, nil, errFQDNFormat
				}

				if isIPv4(ipBytes) {

					var s = net.IP(make([]byte, 16))
					copy(s, mip)
					ret4[string(fqdnBytes)] = s
				} else {

					var s = net.IP(make([]byte, 16))
					copy(s, mip)
					ret6[string(fqdnBytes)] = s
				}

				if line == nil {
					break
				}
			}
			break
		}
		if err != nil {
			break
		}
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		sp = findChar(line, ' ')
		ipBytes, line = line[0:sp], bytes.TrimSpace(line[sp:])
		if mip = toIP(ipBytes); mip == nil {
			// fmt.Println(string(ipBytes))
			return nil, nil, errIPFormat
		}
		if len(line) == 0 {
			err = errFQDNFormat
			break
		}
		for {
			if len(line) == 0 || line[0] == '#' {
				break
			}
			sp = findChar(line, ' ')
			fqdnBytes, line = line[0:sp], bytes.TrimSpace(line[sp:])

			if !testFQDN(fqdnBytes) {
				return nil, nil, errFQDNFormat
			}

			if isIPv4(ipBytes) {

				var s = net.IP(make([]byte, 16))
				copy(s, mip)
				ret4[string(fqdnBytes)] = s
			} else {

				var s = net.IP(make([]byte, 16))
				copy(s, mip)
				ret6[string(fqdnBytes)] = s
			}
		}
	}
	return ret4, ret6, err
}

// read hosts files to map
func LoadHosts() error {
	var err error

	hostspath := config.Config().HostsConfig.HostsPath
	if config.Config().HostsConfig.RelativePath {
		hostspath, err = filepath.Abs(hostspath)
	}
	if err != nil {
		return err
	}

	// HostsIPv4, HostsIPv6, err = Process(config.Hostsfile)
	HostsIPv4, HostsIPv6, err = Process(hostspath)
	return err
}
