package msg

import (
	"errors"
)

type DNSCname struct {
	DNSAnswer
	CNAME []byte
}

func (c *DNSCname) InitFrom(a DNSAnswer) {
	c.Name = a.Name
	c.Type = a.Type
	c.Class = a.Class
	c.RDLength = a.RDLength
	c.RDData = a.RDData
}

func ToDNSCname(a *DNSAnswer) (DNSCname, error) {
	var c DNSCname
	if a.Type != 0x01 {
		return c, errors.New("Resource Record is not a cname type")
	}
	c.InitFrom(*a)
	c.CNAME = a.RDData.([]byte)
	return c, nil
}
