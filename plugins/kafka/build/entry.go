package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"

	"go-sniffer/plugins/kafka/build/internal"
	"go-sniffer/plugins/kafka/build/request"

	"github.com/google/gopacket"
)

const (
	Port    = 9092
	Version = "0.1"
	CmdPort = "-p"
)

type Kafka struct {
	port    int
	version string
	source  map[string]*stream
}

type stream struct {
	packets        chan *packet
	correlationMap map[int32]requestHeader
}

type packet struct {
	isClientFlow bool //客户端->服务器端流
	messageSize  int32

	requestHeader
	responseHeader

	payload io.Reader
}

type requestHeader struct {
	apiKey        int16
	apiVersion    int16
	correlationId int32
	clientId      string
}

type responseHeader struct {
	correlationId int32
}

type messageSet struct {
	offset      int64
	messageSize int32
}

func newMessageSet(r io.Reader) messageSet {
	messageSet := messageSet{}
	messageSet.offset, _ = internal.ReadInt64(r)
	messageSet.messageSize = internal.ReadInt32(r)

	return messageSet
}

type message struct {
	crc        int32
	magicByte  int8
	attributes int8
	key        []byte
	value      []byte
}

var kafka *Kafka
var once sync.Once

func NewInstance() *Kafka {
	once.Do(func() {
		kafka = &Kafka{
			port:    Port,
			version: Version,
			source:  make(map[string]*stream),
		}
	})
	return kafka
}

func (m *Kafka) SetFlag(flg []string) {
	c := len(flg)
	if c == 0 {
		return
	}
	if c>>1 != 1 {
		panic("Kafka 参数数量不正确!")
	}
	for i := 0; i < c; i = i + 2 {
		key := flg[i]
		val := flg[i+1]

		switch key {
		case CmdPort:
			p, err := strconv.Atoi(val)
			if err != nil {
				panic("端口数不正确")
			}
			kafka.port = p
			if p < 0 || p > 65535 {
				panic("参数不正确: 端口范围(0-65535)")
			}
			break
		default:
			panic("参数不正确")
		}
	}
}

func (m *Kafka) BPFFilter() string {
	return "tcp and port " + strconv.Itoa(m.port)
}

func (m *Kafka) Version() string {
	return m.version
}

func (m *Kafka) ResolveStream(net, transport gopacket.Flow, buf io.Reader) {

	//uuid
	uuid := fmt.Sprintf("%v:%v", net.FastHash(), transport.FastHash())

	//resolve packet
	if _, ok := m.source[uuid]; !ok {
		var s = stream{
			packets:        make(chan *packet, 100),
			correlationMap: make(map[int32]requestHeader),
		}
		m.source[uuid] = &s
		go s.resolve()
	}

	//read bi-directional packet
	//server -> client || client -> server
	for {
		p := m.newPacket(net, transport, buf)
		if p == nil {
			continue
		}
		m.source[uuid].packets <- p
	}
}

func (m *Kafka) newPacket(net, transport gopacket.Flow, r io.Reader) *packet {

	//read packet
	p := packet{}

	/*
		bs := make([]byte, 0)
		count, err := r.Read(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("read: %d, buffer: %b", count, bs)
		return nil
	*/

	//read messageSize
	p.messageSize = internal.ReadInt32(r)
	if p.messageSize == 0 {
		return nil
	}
	fmt.Printf("pk.messageSize: %d\n", p.messageSize)

	//set flow direction
	if transport.Src().String() == strconv.Itoa(m.port) {
		p.isClientFlow = false

		respHeader := responseHeader{}
		respHeader.correlationId = internal.ReadInt32(r)
		// TODO: extract request
		p.responseHeader = respHeader

		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, r, int64(p.messageSize-4)); err != nil {
			if err == io.EOF {
				fmt.Println(net, transport, " 关闭")
				return nil
			}
			fmt.Println("流解析错误", net, transport, ":", err)
			return nil
		}

		p.payload = &buf

	} else {
		p.isClientFlow = true

		var clientIdLen = 0
		reqHeader := requestHeader{}
		reqHeader.apiKey = internal.ReadInt16(r)
		reqHeader.apiVersion = internal.ReadInt16(r)
		reqHeader.correlationId = internal.ReadInt32(r)
		reqHeader.clientId, clientIdLen = internal.ReadString(r)
		p.requestHeader = reqHeader
		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, r, int64(p.messageSize-10)-int64(clientIdLen)); err != nil {
			if err == io.EOF {
				fmt.Println(net, transport, " 关闭")
				return nil
			}
			fmt.Println("流解析错误", net, transport, ":", err)
			return nil
		}
		p.payload = &buf
	}

	return &p
}

func (s *stream) resolve() {
	for {
		select {
		case packet := <-s.packets:
			if packet.isClientFlow {
				s.correlationMap[packet.requestHeader.correlationId] = packet.requestHeader
				s.resolveClientPacket(packet)
			} else {
				if _, ok := s.correlationMap[packet.responseHeader.correlationId]; ok {
					s.resolveServerPacket(packet, s.correlationMap[packet.responseHeader.correlationId])
				}
			}
		}
	}
}

func (s *stream) resolveServerPacket(p *packet, rh requestHeader) {
	var msg interface{}
	payload := p.payload

	action := internal.Action{
		Request:    internal.GetRequestName(p.apiKey),
		Direction:  "isServer",
		ApiVersion: p.apiVersion,
	}
	switch int(rh.apiKey) {
	case internal.ProduceRequest:
		msg = request.ReadProduceResponse(payload, rh.apiVersion)
	// case MetadataRequest:
	// 	msg = ReadMetadataResponse(payload, rh.apiVersion)
	default:
		internal.GetRequestName(rh.apiKey)
		fmt.Printf("resolveServerPacket Api: %s TODO: ", internal.GetRequestName(rh.apiKey))
	}

	if msg != nil {
		action.Message = msg
	}
	j, err := json.Marshal(action)
	if err != nil {
		fmt.Printf("resolveServerPacket, Error: %s\n", err.Error())
	}
	fmt.Println(string(j))
}

func (s *stream) resolveClientPacket(p *packet) {
	var msg interface{}
	action := internal.Action{
		Request:    internal.GetRequestName(p.apiKey),
		Direction:  "isClient",
		ApiVersion: p.apiVersion,
	}
	payload := p.payload
	switch int(p.apiKey) {
	case internal.ProduceRequest:
		msg = request.ReadProduceRequest(payload, p.apiVersion)
	// case MetadataRequest:
	// 	msg = ReadMetadataRequest(payload, pk.apiVersion)
	default:
		fmt.Printf("resolveClientPacket Api: %s TODO: ", internal.GetRequestName(p.apiKey))
	}

	if msg != nil {
		action.Message = msg
	}
	j, err := json.Marshal(action)
	if err != nil {
		fmt.Printf("json marshal action failed, err: %s\n", err.Error())
	}
	fmt.Println(string(j))
}
