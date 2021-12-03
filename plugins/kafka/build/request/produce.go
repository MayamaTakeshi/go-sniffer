package request

import (
	"fmt"
	"io"
	"time"

	"go-sniffer/plugins/kafka/build/internal"
)

type Message struct {
	Key          []byte
	Value        []byte
	Offset       int64
	Crc          uint32
	Magic        byte
	CompressCode byte
	Topic        string
	Partition    int32
	TipOffset    int64
}

func ReadMessages(r io.Reader, version int16) []*Message {
	switch version {
	case 0:
		return ReadMessagesV1(r)
	case 1:
		return ReadMessagesV1(r)
	}

	return make([]*Message, 0)
}

func ReadMessagesV1(r io.Reader) []*Message {
	var err error
	messages := make([]*Message, 0)
	for {
		message := Message{}
		message.Offset, err = internal.ReadInt64(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				fmt.Printf("read message offset , err: %+v\n", err)
			}
			break
		}
		_ = internal.ReadInt32(r) // message size
		message.Crc = internal.ReadUint32(r)
		message.Magic = internal.ReadByte(r)
		message.CompressCode = internal.ReadByte(r)
		message.Key = internal.ReadBytes(r)
		message.Value = internal.ReadBytes(r)
		messages = append(messages, &message)
	}
	return messages
}

/**
Produce request Protocol
v0, v1 (supported in 0.9.0 or later) and v2 (supported in 0.10.0 or later)
ProduceRequest => RequiredAcks Timeout [TopicName [Partition MessageSetSize MessageSet]]
	RequiredAcks => int16
	Timeout => int32
	Partition => int32
	MessageSetSize => int32

*/
type ProduceReq struct {
	TransactionalID string
	RequiredAcks    int16
	Timeout         time.Duration
	Topics          []ProduceReqTopic
}

type ProduceReqTopic struct {
	Name       string
	Partitions []ProduceReqPartition
}

type ProduceReqPartition struct {
	ID       int32
	Messages []*Message
}

func ReadProduceRequest(r io.Reader, version int16) *ProduceReq {
	// version == 1

	produceReq := ProduceReq{}

	if int(version) >= internal.ApiV3 {
		produceReq.TransactionalID, _ = internal.ReadString(r)
		fmt.Println(produceReq.TransactionalID)
	}

	produceReq.RequiredAcks = internal.ReadInt16(r)
	produceReq.Timeout = time.Duration(internal.ReadInt32(r)) * time.Millisecond

	l := internal.ReadInt32(r)
	if l > internal.KafkaTopicPartitionMaxCount {
		return &produceReq
	}
	produceReq.Topics = make([]ProduceReqTopic, l)

	for ti := range produceReq.Topics {
		var topic = &produceReq.Topics[ti]
		topic.Name, _ = internal.ReadString(r)

		l := internal.ReadInt32(r)
		topic.Partitions = make([]ProduceReqPartition, l)

		for idx := 0; idx < int(l); idx++ {
			topic.Partitions[idx].ID = internal.ReadInt32(r)
			_ = internal.ReadInt32(r) // partitions size
			topic.Partitions[idx].Messages = ReadMessages(r, version)
		}
	}

	return &produceReq
}

type ProduceResponsePartitions struct {
	PartitionID int32
	Error       int16
	Offset      int64
}

type ProduceResponseTopic struct {
	TopicName    string
	Partitions   []ProduceResponsePartitions
	ThrottleTime int32
}

type ProduceResponse struct {
	Topics []ProduceResponseTopic
}

func ReadProduceResponse(r io.Reader, version int16) *ProduceResponse {
	// version == 1
	produceResponse := ProduceResponse{}
	l := internal.ReadInt32(r)
	produceResponse.Topics = make([]ProduceResponseTopic, 0)
	for i := 0; i < int(l); i++ {
		topic := ProduceResponseTopic{}
		topic.TopicName, _ = internal.ReadString(r)
		pl := internal.ReadInt32(r)
		topic.Partitions = make([]ProduceResponsePartitions, 0)
		for j := 0; j < int(pl); j++ {
			pt := ProduceResponsePartitions{}
			pt.PartitionID = internal.ReadInt32(r)
			pt.Error = internal.ReadInt16(r)
			pt.Offset, _ = internal.ReadInt64(r)
			topic.Partitions = append(topic.Partitions, pt)
		}
		produceResponse.Topics = append(produceResponse.Topics, topic)
	}
	return &produceResponse
}
