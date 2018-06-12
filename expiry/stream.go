package expiry

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// KinesisStream contains all of the functionality used to access Kinesis.
type KinesisStream interface {
	PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
	GetShardIterator(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error)
	GetRecords(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error)
	ListShards(input *kinesis.ListShardsInput) (*kinesis.ListShardsOutput, error)
}

// Stream provides a way to send and receive events using Kinesis.
type Stream struct {
	Name string
	// maxPutSize is the target size of a single HTTP request which pushes data to Kinesis.
	maxPutSize int
	// maxRecordSize is the target size of a Kinesis record. Multiple records are sent in a single request.
	maxRecordSize int
	svc           KinesisStream
}

const (
	// DefaultMaxPutSize is the default target size of a single HTTP request which pushes data to Kinesis.
	DefaultMaxPutSize = 1024 * 1024 * 2 // 2MB
	// DefaultMaxRecordSize is the default target size of a Kinesis record. Multiple records are sent in a single request.
	DefaultMaxRecordSize = 512 * 1024 // 512KB
)

// NewStream creates a new pusher to stream events to Kinesis.
func NewStream(name string) Stream {
	return Stream{
		Name:          name,
		maxRecordSize: DefaultMaxRecordSize,
		maxPutSize:    DefaultMaxPutSize,
		svc:           kinesis.New(session.New()),
	}
}

// Put pushes events onto the stream.
func (p Stream) Put(keys []string) error {
	for _, section := range chunk(keys, p.maxPutSize) { // 1MB per request
		records, err := createPutRecords(section, p.maxRecordSize)
		if err != nil {
			return err
		}
		input := &kinesis.PutRecordsInput{
			StreamName: aws.String(p.Name),
			Records:    records,
		}
		if _, err := p.svc.PutRecords(input); err != nil {
			return err
		}
	}
	return nil
}

func createPutRecords(keys []string, size int) ([]*kinesis.PutRecordsRequestEntry, error) {
	chunks := chunk(keys, size)
	records := make([]*kinesis.PutRecordsRequestEntry, len(chunks))
	for i, sliceOfKeys := range chunks {
		sd := NewStreamData(sliceOfKeys)
		data, err := json.Marshal(sd)
		if err != nil {
			return records, err
		}
		records[i] = &kinesis.PutRecordsRequestEntry{
			PartitionKey: aws.String(createRandomKey()),
			Data:         data,
		}
	}
	return records, nil
}

func createRandomKey() string {
	var vs, op []byte
	rand.Read(vs)
	hex.Encode(op, vs)
	return string(op)
}

func chunk(values []string, maxLength int) (op [][]string) {
	var chunkSize int
	var startIndex, endIndex int
	var v string
	for endIndex, v = range values {
		if chunkSize >= maxLength {
			op = append(op, values[startIndex:endIndex])
			startIndex = endIndex
			chunkSize = 0
		}
		chunkSize += len(v)
	}
	op = append(op, values[startIndex:])
	return
}

// ShardID is the unique ID of a shard.
type ShardID string

// SequenceNumber is the position within the Kinesis stream.
type SequenceNumber string

// StreamPosition stores the reader's position within each shard.
type StreamPosition map[ShardID]SequenceNumber

// Get returns all of the keys added to the stream since the StreamPosition was encountered.
func (p Stream) Get(from StreamPosition) (keys []string, to StreamPosition, err error) {
	shards, err := p.listShards()
	if err != nil {
		err = fmt.Errorf("Get: failed to list all shards: %v", err)
		return
	}
	to = map[ShardID]SequenceNumber{}
	for _, shardID := range shards {
		records, t, read, getRecordsError := p.getRecords(shardID, from[shardID])
		if getRecordsError != nil {
			err = fmt.Errorf("Get: failed to get records: %v", getRecordsError)
			return
		}
		if !read {
			continue
		}
		data, getDataError := getDataFromRecords(records)
		if getDataError != nil {
			err = fmt.Errorf("Get: failed to get data from records: %v", getDataError)
			return
		}
		for _, d := range data {
			keys = append(keys, d.Keys...)
		}
		to[shardID] = t
	}
	return
}

func (p Stream) listShards() (shardIDs []ShardID, err error) {
	var nextToken *string
	var lso *kinesis.ListShardsOutput
	for {
		lso, err = p.svc.ListShards(&kinesis.ListShardsInput{
			StreamName: aws.String(p.Name),
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

func (p Stream) getRecords(shard ShardID, from SequenceNumber) (records []*kinesis.Record, to SequenceNumber, read bool, err error) {
	gsii := &kinesis.GetShardIteratorInput{
		ShardId:           aws.String(string(shard)),
		StreamName:        aws.String(p.Name),
		ShardIteratorType: aws.String("LATEST"),
	}
	if string(from) != "" {
		gsii.ShardIteratorType = aws.String("AFTER_SEQUENCE_NUMBER")
		gsii.StartingSequenceNumber = aws.String(string(from))
	}

	var itr *kinesis.GetShardIteratorOutput
	itr, err = p.svc.GetShardIterator(gsii)
	if err != nil {
		err = fmt.Errorf("Get: failed to get iterator: %v", err)
		return
	}

	si := itr.ShardIterator

	for si != nil {
		var gro *kinesis.GetRecordsOutput
		gro, err = p.svc.GetRecords(&kinesis.GetRecordsInput{ShardIterator: si})
		if err != nil {
			var sis string
			if si != nil {
				sis = *si
			}
			err = fmt.Errorf("Get: failed to get records for shard '%v' with shard iterator '%v': %v", shard, sis, err)
			return
		}
		for _, r := range gro.Records {
			records = append(records, r)
			to = SequenceNumber(*r.SequenceNumber)
			read = true
		}
		si = gro.NextShardIterator
	}
	return
}

func getDataFromRecords(records []*kinesis.Record) (data []StreamData, err error) {
	for _, r := range records {
		var sd StreamData
		err = json.Unmarshal(r.Data, &sd)
		if err != nil {
			return
		}
		data = append(data, sd)
	}
	return
}
