package request

import (
	"io"

	"go-sniffer/plugins/kafka/build/internal"
)

type MetadataReq struct {
	TopicNames []string
}

func ReadMetadataRequest(r io.Reader, version int16) *MetadataReq {
	// version == 0
	metadataReq := MetadataReq{}

	l := internal.ReadInt32(r)
	if l > internal.KafkaTopicPartitionMaxCount {
		return &metadataReq
	}
	for i := 0; i < int(l); i++ {
		topicName, _ := internal.ReadString(r)
		metadataReq.TopicNames = append(metadataReq.TopicNames, topicName)
	}

	return &metadataReq
}

type Broker struct {
	NodeID int32
	Host   string
	Port   int32
}

type PartitionMetada struct {
	ErrorCode      int16
	PartitionIndex int32
	LeaderID       int32
	ReplicaNodes   []int32
	IsrNodes       []int32
}

type TopicMetadata struct {
	ErrorCode  int16
	Name       string
	Partitions []PartitionMetada
}

type MetadataResponse struct {
	Brokers []Broker
	Topics  []TopicMetadata
}

func ReadMetadataResponse(r io.Reader, version int16) *MetadataResponse {
	// version == 0
	metadataResponse := MetadataResponse{}

	// read brokers
	metadataResponse.Brokers = make([]Broker, 0)
	l := internal.ReadInt32(r)
	if l > internal.KafkaTopicPartitionMaxCount {
		return &metadataResponse
	}
	for i := 0; i < int(l); i++ {
		broker := Broker{}
		broker.NodeID = internal.ReadInt32(r)
		broker.Host, _ = internal.ReadString(r)
		broker.Port = internal.ReadInt32(r)
		metadataResponse.Brokers = append(metadataResponse.Brokers, broker)
	}

	// read topics
	metadataResponse.Topics = make([]TopicMetadata, 0)
	l = internal.ReadInt32(r)
	for i := 0; i < int(l); i++ {
		topicMetadata := TopicMetadata{}
		topicMetadata.ErrorCode = internal.ReadInt16(r)
		topicMetadata.Name, _ = internal.ReadString(r)
		pl := internal.ReadInt32(r)
		topicMetadata.Partitions = make([]PartitionMetada, 0)
		for j := 0; j < int(pl); j++ {
			pm := PartitionMetada{}
			pm.ErrorCode = internal.ReadInt16(r)
			pm.PartitionIndex = internal.ReadInt32(r)
			pm.LeaderID = internal.ReadInt32(r)

			pm.ReplicaNodes = make([]int32, 0)
			replicaLen := internal.ReadInt32(r)
			for ri := 0; ri < int(replicaLen); ri++ {
				pm.ReplicaNodes = append(pm.ReplicaNodes, internal.ReadInt32(r))
			}

			pm.IsrNodes = make([]int32, 0)
			isrLen := internal.ReadInt32(r)
			for ri := 0; ri < int(isrLen); ri++ {
				pm.IsrNodes = append(pm.IsrNodes, internal.ReadInt32(r))
			}
			topicMetadata.Partitions = append(topicMetadata.Partitions, pm)
		}
		metadataResponse.Topics = append(metadataResponse.Topics, topicMetadata)
	}

	return &metadataResponse
}
