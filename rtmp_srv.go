package main

import (
        //"bufio"
        "crypto/tls"
        "flag"
        "fmt"
        //"strings"
	"time"
	"encoding/gob"

	//"net"
	"os"
	rtmp "github.com/zhangpeihao/gortmp"
        quicconn "github.com/marten-seemann/quic-conn"
)

var obConn rtmp.OutboundConn
var createStreamChan chan rtmp.OutboundStream
var status uint
var str rtmp.OutboundStream


var (
	url         *string = flag.String("URL", "rtmp://178.62.61.235:1935/show", "The rtmp url to connect.")
	streamName  *string = flag.String("Stream", "stream_name", "Stream name to play.")
	flvFileName *string = flag.String("FLV", "./demo.flv", "FLV file to publishs.")
)

type TestOutboundConnHandler struct {
}

func (handler *TestOutboundConnHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error
	if obConn == nil {
		return
	}
	status, err = obConn.Status()
	fmt.Printf("@@@@@@@@@@@@@status: %d, err: %v\n", status, err)
}

func (handler *TestOutboundConnHandler) OnClosed(conn rtmp.Conn) {
	fmt.Printf("@@@@@@@@@@@@@Closed\n")
}

func (handler *TestOutboundConnHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
}

func (handler *TestOutboundConnHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	fmt.Printf("ReceviedRtmpCommand: %+v\n", command)
}

func (handler *TestOutboundConnHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	fmt.Printf("Stream created: %d\n", stream.ID())
	createStreamChan <- stream
}
func (handler *TestOutboundConnHandler) OnPlayStart(stream rtmp.OutboundStream) {

}
func (handler *TestOutboundConnHandler) OnPublishStart(stream rtmp.OutboundStream) {
	// Set chunk buffer size
	//go publish(stream)
	fmt.Println("Tupe casted")
	str = stream
}

type P struct {
        Buf []byte
        Type uint8
        Timestamp uint32
        AbsoluteTimestamp uint32
}


func main() {
        // utils.SetLogLevel(utils.LogLevelDebug)

        startServer := flag.Bool("s", false, "server")
        flag.Parse()

        if *startServer {
                // start the server
                go func() {
			s := make([]P, 0)
                        certPath := flag.String("certpath", "/etc/letsencrypt/live/streemtechnology.com", "certificate directory")
                        certFile := *certPath + "/fullchain.pem"
                        keyFile := *certPath + "/privkey.pem"
			var err error
                        certs := make([]tls.Certificate, 1)
                        certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
                        if err != nil {
                                panic(err)
                        }
                        config := &tls.Config{
                                Certificates: certs,
                        }

                        ln, err := quicconn.Listen("udp", ":8081", config)
                        if err != nil {
                                panic(err)
                        }
			//rconn, err := net.Dial("tcp", "")
			//if err != nil {
			//	panic(err)	// handle error
			//}

                        fmt.Println("Waiting for incoming connection")
                        conn, err := ln.Accept()
                        if err != nil {
                                panic(err)
                        }
                        fmt.Println("Established connection")

			createStreamChan = make(chan rtmp.OutboundStream)
			testHandler := &TestOutboundConnHandler{}
			fmt.Println("to dial")
			fmt.Println("a")
			obConn, err = rtmp.Dial(*url, testHandler, 100)
			if err != nil {
				fmt.Println("Dial error", err)
				os.Exit(-1)
			}
			fmt.Println("b")
			defer obConn.Close()
			fmt.Println("to connect")
			err = obConn.Connect()
			if err != nil {
				fmt.Printf("Connect error: %s", err.Error())
				os.Exit(-1)
			}
			fmt.Println("c")
			dec := gob.NewDecoder(conn)
			var msg P
                        for {
				select{
				case stream := <-createStreamChan:
					// Publish
					stream.Attach(testHandler)
					err = stream.Publish(*streamName, "live")
					if err != nil {
						fmt.Printf("Publish error: %s", err.Error())
						os.Exit(-1)
					}
                                //message, err := bufio.NewReader(conn).ReadString(0xff)
                                //if err != nil {
                                //        panic(err)
                                //}
                                //fmt.Println("Message from client: ", len(string(message)))
				default:
				if err := dec.Decode(&msg); err != nil {
				    panic(err)
				}
				if str != nil {
					if len(s) >0 && s != nil {
						for i := 0; i < len(s); i++ {
							fmt.Println("sending old packs")
							msg = s[i]
							if err = str.PublishData(msg.Type, msg.Buf, msg.Timestamp); err != nil {
								fmt.Println("PublishData() error:", err)
							}
						}
						s = nil
					}else{
						if err = str.PublishData(msg.Type, msg.Buf, msg.Timestamp); err != nil {
							fmt.Println("PublishData() error:", err)
						}
					}
				}else{
					s = append(s, msg)
					fmt.Println("nope")
				}
				//fmt.Println(msg.Type)
                                // echo back
                                //newmessage := strings.ToUpper(message)
                                //conn.Write([]byte(newmessage + "\n"))
				}
                        }
                }()
        }
	time.Sleep(time.Hour)
}
