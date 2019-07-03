package msg

/*
   +---------------------+
   |        Header       |
   +---------------------+
   |       Question      |
   +---------------------+
   |        Answer       |
   +---------------------+
   |      Authority      |
   +---------------------+
   |      Additional     |
   +---------------------+
*/
type DNSMessage struct {
	Header     DNSHeader
	Question   []DNSQuestion
	Answer     []DNSAnswer
	Authority  []DNSAnswer
	Additional []DNSAnswer

	nameMap map[uint16][]byte
	nameRef uint16
}

/*
map
oa: www
ob: baidu
oc: com
od: \0
oe: oss
of: 指针(值为0x8000|指针ob的值)
og: ssr
oh: sf
oi: cn
oj: \0

nameOffset:
oa  ob    oc  od
www.baidu.com \0

oe  ob    oc  od
oss.baidu.com \0

og  oh oi oj
ssr.sf.cn \0

读取流程:
oa: www
ob: baidu
oc: com
od: \0
-> name = www.baidu.com

oe: oss
of: 0x8000|ob
ob: baidu
oc: com
od: \0
-> name = oss.baidu.com

og: ssr
oh: sf
oi: cn
oj: \0
-> name = ssr.sf.cn


*/
