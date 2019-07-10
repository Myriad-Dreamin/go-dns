package dnssrv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	msg "github.com/Myriad-Dreamin/go-dns/msg"
	mredis "github.com/Myriad-Dreamin/go-dns/redis"
)

type TCPUserRoutine struct {
	*sharedSpace
	Buffer      *bytes.Buffer
	QuitRequest chan bool
	MessageChan chan *bytes.Buffer
	RequestChan chan *bytes.Buffer
	tid         uint16
	readNumber  uint16
}

func NewTCPUserRoutine(
	sharedSpace *sharedSpace,
	tid uint16,
) *TCPUserRoutine {
	return &TCPUserRoutine{
		tid:         tid,
		sharedSpace: sharedSpace,
		Buffer:      sharedSpace.bufferPool.Get().(*bytes.Buffer),
		MessageChan: make(chan *bytes.Buffer),
		RequestChan: make(chan *bytes.Buffer),
		QuitRequest: make(chan bool, 1),
	}
}
func (rt *TCPUserRoutine) RequestQuit() bool {
	if rt == nil {
		return false
	}
	rt.QuitRequest <- true
	return true
}

func (rt *TCPUserRoutine) Run() {
	if rt == nil {
		return
	}
	var n int
	var err error
	for {
		select {
		case <-rt.QuitRequest:
			rt.quit <- true
			fmt.Println("TCPUserRoutine")
			return
		default:
			var tcpConn *net.TCPConn
			// accept a net tcp connection
			rt.dispatcher.connTCP.SetDeadline(time.Now().Add(1 * time.Second))
			{
				for len(rt.MessageChan) > 0 {
					rt.bufferPool.Put(<-rt.MessageChan)
				}
				tcpConn, err = rt.dispatcher.connTCP.AcceptTCP()
				if err != nil {
					if er, ok := err.(net.Error); !ok {
						rt.logger.Errorf("accept error: %v", err)
					} else if er.Timeout() {
						// continue
					} else {
						rt.logger.Errorf("failed when accepting tcp connection, error: %v", err)
					}
					continue
				}
			}
			rt.logger.Infoln("routine", rt.tid, "getting", tcpConn.RemoteAddr())

			rt.dispatcher.connTCP.SetDeadline(time.Now().Add(1 * time.Second))

			// preparing
			tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))
			rt.Buffer.Reset()
			var b, c []byte
			b = rt.bytesPool.Get().([]byte)
			rt.readNumber = 0
			conn := mredis.RedisCacheClient.Pool.Get()
			for {
				// accept a dns message
				for {
					if rt.Buffer.Len() > 1 && rt.readNumber == 0 {
						binary.Read(rt.Buffer, binary.BigEndian, &rt.readNumber)
					}
					if rt.readNumber != 0 && rt.Buffer.Len() >= int(rt.readNumber) {
						c = rt.Buffer.Next(int(rt.readNumber))
						break
					}

					tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))
					n, err = tcpConn.Read(b)
					if err != nil {
						if err != io.EOF {
							rt.logger.Errorf("failed when reading tcp flow, error: %v", err)
						}
						rt.logger.Errorf("failed when reading tcp flow, error: %v", err)
						rt.readNumber = 0
						break
					}
					fmt.Println(rt.tid, "buffer cap", len(b), rt.Buffer.Cap())
					fmt.Println("?", rt.readNumber, b[0:n])
					if n == 0 {
						continue
					}
					tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))
					_, err = rt.Buffer.Write(b[0:n])
					if err != nil {
						rt.logger.Errorf("buffering error: %v", err)
						rt.Buffer.Reset()
						rt.readNumber = 0
						break
					}
				}
				// the message is stored in c([]byte)
				// if read a bad mesage, the rt.readNumber will be reset to 0
				// otherwise the rt.readNumber will be equal to len(message)

				// flush buffer
				if rt.Buffer.Cap() < 600 {
					var t = rt.bufferPool.Get().(*bytes.Buffer)
					_, err = rt.Buffer.WriteTo(t)
					if err != nil {
						rt.logger.Errorf("convert buffering error: %v", err)
						rt.bufferPool.Put(t)
						continue
					} else {
						rt.Buffer.Reset()
						rt.bufferPool.Put(rt.Buffer)
						rt.Buffer = t
					}
				}

				// read bad message
				if rt.readNumber == 0 {
					rt.logger.Errorf("aborted %v", rt.tid)
					goto reset_and_reaccept_new_link
				}

				// convert message
				var message msg.DNSMessage
				_, err = message.Read(c)
				if err != nil {
					rt.logger.Errorf("failed when decoding tcp flow, error: %v", err)
					tcpConn.Close()
					conn.Close()
					break
				}
				rt.logger.Infof("new message(tcp) incoming: id, address: %v, %v", message.Header.ID, tcpConn.RemoteAddr())
				message.Print()

				// reply the questions
				reply := msg.NewDNSMessageReply(message.Header.ID, message.Header.Flags, message.Question)
				if mredis.FindCache(reply, conn) {
					// reply.Print()
					b, err := reply.CompressToBytes()
					if err != nil {
						rt.logger.Errorf("get redis cache error: %v", err)
						tcpConn.Close()
						conn.Close()
						break
					}

					tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))

					var lenb = uint16(len(b))
					fmt.Println(lenb, b)
					err = binary.Write(tcpConn, binary.BigEndian, &lenb)
					if err != nil {
						rt.logger.Errorf("write to client error: %v", err)
						goto reset_and_reaccept_new_link
					}

					_, err = tcpConn.Write(b)
					if err != nil {
						rt.logger.Errorf("write to client error: %v", err)
						goto reset_and_reaccept_new_link
					}

					rt.logger.Infof("using redis cache reply to address: %v, %v", message.Header.ID, tcpConn.RemoteAddr())
				} else {
					// send message to quest remote server
					fid := message.Header.ID
					message.Header.ID = rt.tid
					b, err = message.CompressToBytes()
					if err != nil {
						rt.logger.Errorf("convert error: %v", err)
						goto reset_and_reaccept_new_link
					}

					var buf = rt.bufferPool.Get().(*bytes.Buffer)
					_, err = buf.Write(b)
					if err != nil {
						rt.logger.Errorf("failed when encoding tcp message, error: %v", err)
						rt.bufferPool.Put(buf)
						goto reset_and_reaccept_new_link
					}

					tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))

					// clear message channel
					for len(rt.MessageChan) > 0 {
						rt.bufferPool.Put(<-rt.MessageChan)
					}

					rt.logger.Infoln("questing remote server", rt.tid)
					rt.dispatcher.messageChan <- buf
					select {
					case buf = <-rt.MessageChan:
					case <-time.After(time.Second * 10):
						//todo: reply...
						rt.logger.Errorf("timeout: routine name, %v", rt.tid)
						goto reset_and_reaccept_new_link
					}

					tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))
					_, err = message.Read(buf.Bytes())
					buf.Reset()
					rt.bufferPool.Put(buf)
					if err != nil {
						rt.logger.Errorf("read error: %v", err)
						goto reset_and_reaccept_new_link
					}
					// if tid != message.Header.ID {
					// 	rt.logger.Errorf("not matching..., serving %v", servingAddr)
					// }
					// message.Print()
					message.Header.ID = fid

					mredis.MessageToRedis(message, conn)

					b, err = message.CompressToBytes()

					// b[0] = byte(fid >> 8)
					// b[1] = byte(fid & 0xff)
					if err != nil {
						rt.logger.Errorf("convert error: %v", err)
						goto reset_and_reaccept_new_link
					}
					tcpConn.SetDeadline(time.Now().Add(rt.tcpTimeout))

					var lenb = uint16(len(b))
					fmt.Println(lenb, b)
					err = binary.Write(tcpConn, binary.BigEndian, &lenb)
					if err != nil {
						rt.logger.Errorf("write to client error: %v", err)
						goto reset_and_reaccept_new_link
					}

					_, err = tcpConn.Write(b)
					if err != nil {
						rt.logger.Errorf("write to client error: %v", err)
						goto reset_and_reaccept_new_link
					}

					rt.logger.Infof("reply to address: %v, %v", message.Header.ID, tcpConn.RemoteAddr())
				}
				// case idx := <-rt.TCPRoutineLimit:
				// 	go rt.ServeTCPFromOut(idx, rt.TCPBuffer[idx-UDPRange])
				// case idx := <-rt.TCPWriteRoutineLimit:
				// 	go rt.ServeTCPWriteToOut(idx)
				// case idx := <-rt.TCPReadRoutineLimit:
				// 	go rt.ServeTCPReadFromOut(idx, rt.TCPReadBuffer[idx-UDPRange])
			}
			continue
		reset_and_reaccept_new_link:
			tcpConn.Close()
			conn.Close()
		}
	}
}
