package hosts

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"regexp"
)

var (
	errTooMuchItem = errors.New("too much item in a line")
	errIPFormat    = errors.New("bad format of ip")
	errFQDNFormat  = errors.New("bad format of domain name")
)

type HostMapping = map[string][]byte

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

// test0to255 = "[0:9]|[1:9][0:9]|1[0:9]{2}|2[0:4][0:9]|25[0:5]"
// testipv4   = "^" + test0to255 + `\.` + test0to255 + `\.` + test0to255 + `\.` + test0to255 + "$"
)

var (
	regFQDN = regexp.MustCompile("^(([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z])))|((([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z]))\\.)*)([0-9A-Za-z](()|([-0-9A-Za-z]{0,60}[0-9A-Za-z])))))$")
	// rega    = regexp.MustCompile(test63NRegExp)
	// regs    = regexp.MustCompile("^(([0-9A-Za-z](()|([-0-9A-Za-z]{0,61}[0-9A-Za-z]))\\.)*)([0-9A-Za-z](()|([-0-9A-Za-z]{0,60}[0-9A-Za-z])))$")
)

func testIP(ipaddr []byte) bool {
	return net.ParseIP(string(ipaddr)) != nil
}

func testFQDN(fqdn []byte) bool {
	return regFQDN.Match(fqdn)
}

// func testFQDNx(fqdn []byte) bool {
// 	return regs.Match(fqdn) || rega.Match(fqdn)
// }

func Process(filePath string) (HostMapping, error) {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	// var buf = bytes.NewBuffer(make([]byte, 0, 1024))
	var sp int
	var ipBytes, fqdnBytes []byte
	var ret = make(HostMapping)
	for {
		line, err := rd.ReadBytes('\n')
		if err != nil || io.EOF == err {
			break
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if line[0] == '#' {
			continue
		}

		sp = findChar(line, ' ')
		ipBytes, line = line[0:sp], bytes.TrimSpace(line[sp:])
		sp = findChar(line, ' ')
		fqdnBytes, line = line[0:sp], bytes.TrimSpace(line[sp:])
		if len(line) != 0 && line[0] != '#' {
			return nil, errTooMuchItem
		}

		if !testIP(ipBytes) {
			return nil, errIPFormat
		}
		if !testFQDN(fqdnBytes) {
			return nil, errFQDNFormat
		}
		ret[string(fqdnBytes)] = ipBytes
	}
	return ret, nil
}
