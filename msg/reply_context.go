package msg

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	rtype "github.com/Myriad-Dreamin/go-dns/msg/rec/rtype"
	mdnet "github.com/Myriad-Dreamin/go-dns/net"
)

type ReplyContext struct {
	Buf       *mdnet.IO
	SuffixMap map[string]int
	Message   *DNSMessage
}

var s mdnet.IO

func NewReplyContext(dnm *DNSMessage) (rc *ReplyContext) {
	return &ReplyContext{
		Buf:       mdnet.NewIO(),
		SuffixMap: make(map[string]int),
		Message:   dnm,
	}
}

// 3ebd00b184dbb2acd4b21f1c0800450000dd5a6600007f110b47c0a82a09c0a82a090035ce6900c9c406

// 0002818000010002000100010763617074636861056774696d6703636f6d00001c0001
// adde818000010002000100010763617074636861056774696d6703636f6d00001c0001
// Name c00c 0005 0001 0000012c 0020 0763617074636861056774696d6703636f6d05636c6f7564027463027171c01ac02f00050001000000bb000b03703231047463646ec04ac05f0006000100000108002a076e732d63646e31c04a097765626d6173746572c04a4fd6eede0000012c00000258000151800000012c000029100000000000001c000a0018d13548243231cf727ee4e34e5d1f2fdfdc2ddf657e74aa90
// Name c00c 0005 0001 0000012c 0100 0763617074636861056774696d6703636f6d05636c6f7564027463027171c01a00050001000000bb0100047463646ec04a0006000100000108002a076e732d63646e31c04a097765626d6173746572c04a4fd6eede0000012c00000258000151800000012c000029100000000000001c000a0018d13548243231cf727ee4e34e5d1f2fdfdc2ddf657e74aa90
// 7f36818000010002000100010763617074636861056774696d6703636f6d00001c0001
//      c00c 0005 0001 0000004f 200763617074636861056774696d6703636f6d05636c6f7564027463027171c01ac02d00050001000002050b03703231047463646ec048c05c00060001000000e4002a076e732d63646e31c04a097765626d6173746572c04a4fd6eede0000012c00000258000151800000012c000029100000000000001c000a001836c54af6acd2d76dc5ae10fb5d1f35f3d31783c8bed4d8a4
func (ctx *ReplyContext) CompressName(bytename []byte, reserveSpace int) ([]byte, error) {
	name := strings.Split(string(bytename), ".")
	var (
		buf    = new(bytes.Buffer)
		suffix string
		trunc  int
		nxoff  int
		offset int
		prelen int
		flag   bool
	)
	fmt.Println(ctx.SuffixMap)
	trunc = len(name)

	if trunc == 1 {
		return []byte{0}, nil
	}

	suffix = name[len(name)-1]
	offset = ctx.Buf.Len()
	prelen = len(bytename) + reserveSpace
	for j := len(name) - 1; j >= 0; j-- {
		if j == len(name)-1 {
			suffix = name[j]
			prelen -= len(name[j])
		} else {
			suffix = name[j] + "." + suffix
			prelen -= len(name[j]) + 1
		}
		if _, ok := ctx.SuffixMap[suffix]; ok == false {
			fmt.Println(suffix)
			ctx.SuffixMap[suffix] = offset + prelen
		} else {
			flag = true
			trunc = j
			nxoff = ctx.SuffixMap[suffix]
		}
	}
	fmt.Println(trunc)
	if flag {
		for j := 0; j < trunc; j++ {
			fmt.Println(len(name[j]))
			fmt.Printf("%x\n", byte(len(name[j])))
			buf.WriteByte(byte(len(name[j])))
			buf.Write([]byte(name[j]))
		}
		tmp := make([]byte, 2)
		if nxoff > 0x3fff {
			return nil, errors.New("Offset out of range")
		}
		tmp[0] = uint8(0xc0 | (nxoff>>8)&0xff)
		tmp[1] = uint8(nxoff & 0xff)
		buf.Write(tmp)
	} else {
		for j := 0; j < len(name); j++ {
			fmt.Println(len(name[j]))
			fmt.Printf("%x\n", byte(len(name[j])))
			buf.WriteByte(byte(len(name[j])))
			buf.Write([]byte(name[j]))
		}
		buf.WriteByte(byte(0))
	}
	fmt.Println(ctx.SuffixMap)
	return buf.Bytes(), nil
}

func (ctx *ReplyContext) InsertName(b []byte) (err error) {
	if b, err = ctx.CompressName(b, 0); err != nil {
		return
	}
	fmt.Println("getting name", b)
	ctx.Buf.Write(b)
	return
}

func (ctx *ReplyContext) InsertNameWithLength(b []byte) (err error) {
	if b, err = ctx.CompressName(b, 2); err != nil {
		return
	}
	fmt.Println("getting compressing name", uint16(len(b)), b)
	ctx.Buf.Write(uint16(len(b)))
	ctx.Buf.Write(b)
	return
}

func (ctx *ReplyContext) InsertQuestion(q DNSQuestion) (err error) {
	if err = ctx.InsertName(q.Name); err != nil {
		return
	}
	ctx.Buf.Write(q.Type)
	ctx.Buf.Write(q.Class)
	return
}

func (ctx *ReplyContext) InsertAnswer(a DNSAnswer) (err error) {
	if err = ctx.InsertName(a.Name); err != nil {
		return
	}
	ctx.Buf.Write(a.Type)
	ctx.Buf.Write(a.Class)
	ctx.Buf.Write(a.TTL)
	/*
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
	*/
	// TODO: Test a.Type

	if a.Type == rtype.CNAME {
		fmt.Println("compressing cname", string(a.RDData))
		if err = ctx.InsertNameWithLength(a.RDData); err != nil {
			return
		}
	} else {
		ctx.Buf.Write(a.RDLength)
		ctx.Buf.Write(a.RDData)
	}
	return
}

func (ctx *ReplyContext) Bytes() (b []byte, err error) {
	ctx.Buf.Write(ctx.Message.Header)

	for _, que := range ctx.Message.Question {
		if err = ctx.InsertQuestion(que); err != nil {
			return nil, err
		}
	}

	for _, ans := range ctx.Message.Answer {
		if err = ctx.InsertAnswer(ans); err != nil {
			return nil, err
		}
	}

	for _, auth := range ctx.Message.Authority {
		if err = ctx.InsertAnswer(auth); err != nil {
			return nil, err
		}
	}

	for _, add := range ctx.Message.Additional {
		if err = ctx.InsertAnswer(add); err != nil {
			return nil, err
		}
	}

	return ctx.Buf.Bytes(), nil
}
