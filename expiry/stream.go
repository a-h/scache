package stream

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// Pusher provides a way to push events to Kinesis.
type Pusher struct {
	StreamName string
}

// NewPusher creates a new pusher to stream events to Kinesis.
func NewPusher(streamName string) Pusher {
	return Pusher{
		StreamName: streamName,
	}
}

// Push pushes events to the stream.
func (p Pusher) Push(keys []string) (err error) {
	svc := kinesis.New(session.New())
	records, err := getRecords(keys)
	if err != nil {
		return
	}
	input := &kinesis.PutRecordsInput{
		StreamName: aws.String(p.StreamName),
		Records:    records,
	}
	_, err = svc.PutRecords(input)
	return
}

func getRecords(keys []string) ([]*kinesis.PutRecordsRequestEntry, error) {
	records := make([]*kinesis.PutRecordsRequestEntry, len(keys))
	// Chunk into 512KB messages, Kinesis's maximum is 1MB.
	for i, sliceOfKeys := range chunk(keys, 512*1024) {
		data, err := json.Marshal(sliceOfKeys)
		if err != nil {
			return records, err
		}
		records[i] = &kinesis.PutRecordsRequestEntry{
			PartitionKey: aws.String(randomKey()),
			Data:         data,
		}
	}
	return records, nil
}

func randomKey() string {
	var vs, op []byte
	rand.Read(vs)
	hex.Encode(op, vs)
	return string(op)
}

func chunk(values []string, maxLength int) (op [][]string) {
	var chunkSize int
	var startIndex, endIndex int
	for endIndex, v := range values {
		chunkSize += len(v)
		if chunkSize > maxLength {
			op = append(op, values[startIndex:endIndex])
			startIndex = endIndex
			chunkSize = 0
		}
	}
	if startIndex != endIndex {
		op = append(op, values[startIndex:endIndex])
	}
	return
}

type ShardID string
type SequenceNumber string
type StreamPosition map[ShardID]SequenceNumber

func (p Pusher) Get(from StreamPosition) (keys []string, to StreamPosition, err error) {
	shards, err := p.ListAllShards()
	if err != nil {
		err = fmt.Errorf("Get: failed to list all shards: %v", err)
		return
	}
	for _, shardID := range shards {
		//TODO: Check that the sequence number date makes sense. If we haven't got anything in our cache, we don't care.
		//TODO: We only care about changes made to data equal to or after the oldest date data currently in the cache.
		shardKeys, t, read, getRecordsError := p.GetAllRecords(shardID, from[shardID])
		if getRecordsError != nil {
			err = fmt.Errorf("Get: failed to get records: %v", getRecordsError)
			return
		}
		if !read {
			continue
		}
		keys = append(keys, shardKeys...)
		to[shardID] = t
	}
	return
}

func (p Pusher) GetAllRecords(shard ShardID, from SequenceNumber) (keys []string, to SequenceNumber, read bool, err error) {
	svc := kinesis.New(session.New())

	gsii := &kinesis.GetShardIteratorInput{
		ShardId:    aws.String(string(shard)),
		StreamName: aws.String(p.StreamName),
	}
	if string(from) != "" {
		gsii.ShardIteratorType = aws.String("AFTER_SEQUENCE_NUMBER")
		gsii.StartingSequenceNumber = aws.String(string(from))
	}

	var itr *kinesis.GetShardIteratorOutput
	itr, err = svc.GetShardIterator(gsii)
	if err != nil {
		err = fmt.Errorf("Get: failed to get iterator: %v", err)
		return
	}

	si := itr.ShardIterator

	for si != nil {
		var gro *kinesis.GetRecordsOutput
		gro, err = svc.GetRecords(&kinesis.GetRecordsInput{ShardIterator: si})
		if err != nil {
			err = fmt.Errorf("Get: failed to get records for shard '%v' with shard iterator '%v': %v", shardID, si, err)
			return
		}
		for _, r := range gro.Records {
			var keySlice []string
			err = json.Unmarshal(r.Data, &keySlice)
			if err != nil {
				err = fmt.Errorf("Get: failed to get unmarshal data for record '%v' for shard '%v' with shard iterator '%v': %v", r.SequenceNumber, shardID, itr.ShardIterator, err)
				return
			}
			keys = append(keys, keySlice...)
			to = SequenceNumber(*r.SequenceNumber)
		}
		si = gro.NextShardIterator
	}
	return
}

func (p Pusher) ListAllShards() (shardIDs []ShardID, err error) {
	svc := kinesis.New(session.New())
	var nextToken *string
	var lso *kinesis.ListShardsOutput
	for {
		lso, err = svc.ListShards(&kinesis.ListShardsInput{
			StreamName: aws.String(p.StreamName),
			MaxResults: aws.Int64(1000),
			NextToken:  nextToken,
		})
		if err != nil {
			return
		}
		for _, s := range lso.Shards {
			shardIDs = append(shardIDs, ShardID(*s.ShardId))
		}
		if lso.NextToken == nil {
			break
		}
		nextToken = lso.NextToken
	}
	return
}
