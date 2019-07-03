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
}
