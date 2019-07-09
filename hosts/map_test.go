package hosts

import (
	"fmt"
	"testing"
)

func TestReg(t *testing.T) {
	var b bool
	// b = testFQDNx([]byte("www.a"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("www.a-"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("www.a3"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("www.a.com"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("a"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("2"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("a3.a2.333"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("a-.a2.333"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("aa..333"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	// fmt.Println(b)
	// b = testFQDNx([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	// fmt.Println(b)
	//
	// fmt.Println("")

	b = testFQDN([]byte("www.a"))
	fmt.Println(b)
	b = testFQDN([]byte("www.a-"))
	fmt.Println(b)
	b = testFQDN([]byte("www.a3"))
	fmt.Println(b)
	b = testFQDN([]byte("www.a.com"))
	fmt.Println(b)
	b = testFQDN([]byte("a"))
	fmt.Println(b)
	b = testFQDN([]byte("2"))
	fmt.Println(b)
	b = testFQDN([]byte("a3.a2.333"))
	fmt.Println(b)
	b = testFQDN([]byte("a-.a2.333"))
	fmt.Println(b)
	b = testFQDN([]byte("aa..333"))
	fmt.Println(b)
	b = testFQDN([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	fmt.Println(b)
	b = testFQDN([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	fmt.Println(b)
	b = testFQDN([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	fmt.Println(b)
	b = testFQDN([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	fmt.Println(b)
}

func TestRead(t *testing.T) {
	fmt.Println(Process("test.txt"))
}
func TestRead2(t *testing.T) {
	_, err := Process("dnsrelay.txt")
	fmt.Println(err)
}
