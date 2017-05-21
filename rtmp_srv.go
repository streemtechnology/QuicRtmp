package main

import (
	//"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	//"strings"
	"encoding/gob"
	"time"

	//"net"
	quicconn "github.com/marten-seemann/quic-conn"
	rtmp "github.com/zhangpeihao/gortmp"
	"os"
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
	Buf               []byte
	Type              uint8
	Timestamp         uint32
	AbsoluteTimestamp uint32
}

func main() {
	// utils.SetLogLevel(utils.LogLevelDebug)

	startServer := flag.Bool("s", false, "server")
	flag.Parse()

	if *startServer {
		// start the server
		go func() {
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
			stream := <-createStreamChan
			stream.Attach(testHandler)
			err = stream.Publish(*streamName, "live")
			if err != nil {
				fmt.Printf("Publish error: %s", err.Error())
				os.Exit(-1)
			}
			startTs := uint32(0)
			diff1 := uint32(0)
			for {
				var msg P
				if err := dec.Decode(&msg); err != nil {
					fmt.Println(err)
					panic(err)
				}
				if str != nil {
					if startTs == uint32(0) {
						startTs = msg.Timestamp
					}
					if msg.Timestamp+diff1 > diff1 {
						diff1 = msg.Timestamp + diff1
					}
					fmt.Println(msg.Type, len(msg.Buf), msg.Timestamp, diff1)
					if err = str.PublishData(msg.Type, msg.Buf, 0); err != nil {
						fmt.Println("PublishData() error:", err)
					}
				} else {
					fmt.Println("nope can't publish", msg)
				}
			}
		}()
	}
	time.Sleep(time.Hour)
}
