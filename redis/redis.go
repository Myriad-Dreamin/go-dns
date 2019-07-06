package mredis

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Myriad-Dreamin/go-dns/msg"
	qtype "github.com/Myriad-Dreamin/go-dns/msg/rec/qtype"
	rtype "github.com/Myriad-Dreamin/go-dns/msg/rec/rtype"
	"github.com/garyburd/redigo/redis"
	"strings"
)

func AnswersToRedis(answers []msg.DNSAnswer, conn redis.Conn) (int, error) {
	var cnt int
	for _, ans := range answers {
		// key, err := ans.RedisRandomKey()
		key, err := ans.RedisHashKey()
		if err != nil {
			return 0, err
		}
		//b := ans.RDData
		b, err := ans.ToBytes()
		if err != nil {
			return 0, err
		}
		// conn.Do("set", key, b, "EX", ans.TTL)
		switch ans.Type {
		case rtype.A, rtype.NS, rtype.CNAME, rtype.AAAA:
			conn.Send("set", key, b, "EX", ans.TTL)
			cnt += 1
		default:
			return 0, errors.New("Type it not suppoted by redis")
		}

	}
	return cnt, nil
}

func AuthorityToRedis(que msg.DNSQuestion, answers []msg.DNSAnswer, conn redis.Conn) (int, error) {
	var cnt int
	for _, ans := range answers {
		key, err := ans.RedisAuthorityHashKey(que.Type)
		if err != nil {
			return 0, err
		}
		ans.ToFormalSOA()
		b, err := ans.ToBytes()
		if err != nil {
			return 0, err
		}
		// conn.Do("set", key, b, "EX", ans.TTL)
		switch ans.Type {
		case rtype.SOA:
			conn.Send("set", key, b, "EX", ans.RDData.(*msg.SOA).MinimumTTL)
			cnt += 1
		default:
			return 0, errors.New("Type it not suppoted by redis")
		}
		if !bytes.Equal(que.Name, ans.Name) {
			ans.Name = que.Name
			key, err := ans.RedisAuthorityHashKey(que.Type)
			if err != nil {
				return 0, err
			}
			//b := ans.RDData
			b, err := ans.ToBytes()
			if err != nil {
				return 0, err
			}
			// conn.Do("set", key, b, "EX", ans.TTL)
			switch ans.Type {
			case rtype.SOA:
				conn.Send("set", key, b, "EX", ans.RDData.(*msg.SOA).MinimumTTL)
				cnt += 1
			default:
				return 0, errors.New("Type it not suppoted by redis")
			}
		}

	}
	return cnt, nil
}

func MessageToRedis(msg msg.DNSMessage, conn redis.Conn) error {
	var (
		total int
		n     int
		err   error
	)
	if n, err = AnswersToRedis(msg.Answer, conn); err != nil {
		return err
	}
	total += n
	if n, err = AuthorityToRedis(msg.Question[0], msg.Authority, conn); err != nil {
		return err
	}
	total += n
	if n, err = AnswersToRedis(msg.Additional, conn); err != nil {
		return err
	}
	total += n
	conn.Flush()
	for i := 0; i < total; i++ {
		if _, err := conn.Receive(); err != nil {
			return err
		}
	}
	return nil
}

func HasRecord(keys []string, prefix string) int {
	for i, str := range keys {
		if strings.HasPrefix(str, prefix) {
			return i
		}
	}
	return -1
}

func FindCache(m *msg.DNSMessage, conn redis.Conn) bool {
	var (
		replyans msg.DNSAnswer
	)
	for _, que := range m.Question {
		domain := string(que.Name)
		searchkey, err := que.RedisKey()
		if err != nil {
			return false
		}
		switch que.Type {
		case qtype.A, qtype.AAAA:
			keys, err := redis.Strings(conn.Do("keys", searchkey+":*"))
			if err != nil {
				return false
			}
			if len(keys) == 0 { // Find CNAME
				keys, err = redis.Strings(conn.Do("keys", domain+":CNAME:*"))
				if len(keys) > 1 {
					fmt.Print("Multiple CNAME error")
					return false
				} else if len(keys) != 0 {
					key := keys[0]
					bs, err := redis.Bytes(conn.Do("get", key))
					if err != nil {
						return false
					}
					replyans.ReadFrom(bs, 0)
					m.InsertAnswer(replyans)
					searchkey, err = replyans.RedisKey()
					if err != nil {
						return false
					}
					keys, err = redis.Strings(conn.Do("keys", string(replyans.Name)+":CNAME"))
					if err != nil {
						return false
					}
				}
			}
			if len(keys) == 0 { // Find SOA
				keys, err = redis.Strings(conn.Do("keys", domain+":SOA:"+msg.Typename[que.Type]+"*"))
				if len(keys) == 0 {
					return false
				} else {
					return AddAuthority(m, keys, conn)
				}
			}
			return AddAnswer(m, keys, conn)
		case qtype.NS, qtype.CNAME:
			keys, err := redis.Strings(conn.Do("keys", searchkey+":*"))
			if err != nil {
				return false
			}
			return AddAnswer(m, keys, conn)
		default:
			return false
		}
	}
	return true
}

func AddAnswer(m *msg.DNSMessage, keys []string, conn redis.Conn) bool {
	if len(keys) == 0 {
		return false
	}
	var replyans msg.DNSAnswer
	for _, k := range keys {
		conn.Send("get", k)
		conn.Send("ttl", k)
	}
	conn.Flush()
	for i := 0; i < len(keys); i++ {
		res, err := conn.Receive()
		ttl, err := conn.Receive()
		if err != nil {
			return false
		}
		replyans.ReadFrom(res.([]byte), 0)
		replyans.SetTTL(uint32(ttl.(int64)))
		m.InsertAnswer(replyans)
	}
	return true
}

func AddAuthority(m *msg.DNSMessage, keys []string, conn redis.Conn) bool {
	if len(keys) == 0 {
		return false
	}
	var replyans msg.DNSAnswer
	for _, k := range keys {
		conn.Send("get", k)
		conn.Send("ttl", k)
	}
	conn.Flush()
	for i := 0; i < len(keys); i++ {
		res, err := conn.Receive()
		ttl, err := conn.Receive()
		if err != nil {
			return false
		}
		replyans.ReadFrom(res.([]byte), 0)
		replyans.SetTTL(uint32(ttl.(int64)))
		m.InsertAuthority(replyans)
	}
	return true
}
