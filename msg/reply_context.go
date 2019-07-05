package msg

import (
	"errors"
	"fmt"
	"strings"

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

func (ctx *ReplyContext) InsertName(bytename []byte) error {
	name := strings.Split(string(bytename), ".")
	var (
		suffix string
		trunc  int
		nxoff  int
		offset int
		prelen int
		flag   bool
	)
	// fmt.Println(ctx.SuffixMap)
	suffix = name[len(name)-1]
	// trunc = len(name)
	offset = ctx.Buf.Len()
	prelen = len(bytename)
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
	// fmt.Println(trunc)
	if flag {
		for j := 0; j < trunc; j++ {
			// fmt.Println(len(name[j]))
			// fmt.Printf("%x\n", byte(len(name[j])))
			ctx.Buf.Write(byte(len(name[j])))
			ctx.Buf.Write([]byte(name[j]))
		}
		tmp := make([]byte, 2)
		if nxoff > 0x3fff {
			return errors.New("Offset out of range")
		}
		tmp[0] = uint8(0xc0 | (nxoff>>8)&0xff)
		tmp[1] = uint8(nxoff & 0xff)
		ctx.Buf.Write(tmp)
	} else {
		for j := 0; j < len(name); j++ {
			fmt.Println(len(name[j]))
			fmt.Printf("%x\n", byte(len(name[j])))
			ctx.Buf.Write(byte(len(name[j])))
			ctx.Buf.Write([]byte(name[j]))
		}
		ctx.Buf.Write(byte(0))
	}
	fmt.Println(ctx.SuffixMap)
	return nil
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
	ctx.Buf.Write(a.RDLength)

	// TODO: Test a.Type
	ctx.Buf.Write(a.RDData)
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
