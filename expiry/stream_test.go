package expiry

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

type TestKinesisStream struct {
	PutRecordsFunc       func(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
	ListShardsFunc       func(input *kinesis.ListShardsInput) (*kinesis.ListShardsOutput, error)
	GetShardIteratorFunc func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error)
	GetRecordsFunc       func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error)
}

func (tks TestKinesisStream) PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {
	return tks.PutRecordsFunc(input)
}
func (tks TestKinesisStream) ListShards(input *kinesis.ListShardsInput) (*kinesis.ListShardsOutput, error) {
	return tks.ListShardsFunc(input)
}
func (tks TestKinesisStream) GetShardIterator(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
	return tks.GetShardIteratorFunc(input)
}
func (tks TestKinesisStream) GetRecords(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
	return tks.GetRecordsFunc(input)
}

func DefaultListShardsFunc(input *kinesis.ListShardsInput) (op *kinesis.ListShardsOutput, err error) {
	if input.NextToken == nil {
		op = &kinesis.ListShardsOutput{
			NextToken: aws.String("shard_list_1"),
			Shards: []*kinesis.Shard{
				{
					ShardId: aws.String("shard_1"),
				},
				{
					ShardId: aws.String("shard_2"),
				},
			},
		}
		return
	}
	switch *input.NextToken {
	case "shard_list_1":
		op = &kinesis.ListShardsOutput{
			NextToken: nil,
			Shards: []*kinesis.Shard{
				{
					ShardId: aws.String("shard_3"),
				},
				{
					ShardId: aws.String("shard_4"),
				},
			},
		}
		return
	default:
		err = fmt.Errorf("unexpected next shard token: %v", input.NextToken)
	}
	return
}

func TestListShards(t *testing.T) {
	s := NewStream("test")
	s.svc = TestKinesisStream{
		ListShardsFunc: DefaultListShardsFunc,
	}
	ids, err := s.listShards()
	if err != nil {
		t.Errorf("unexepected error listing shards: %v", err)
	}
	if !reflect.DeepEqual(ids, []ShardID{"shard_1", "shard_2", "shard_3", "shard_4"}) {
		t.Errorf("unexpected list of shards: %v", ids)
	}
}

func TestGetRecords(t *testing.T) {
	tests := []struct {
		name                 string
		getRecordsFunc       func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error)
		getShardIteratorFunc func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error)
		from                 SequenceNumber
		expectedRead         bool
		expectedRecordCount  int
		expectedTo           SequenceNumber
	}{
		{
			name: "get single set of records",
			getShardIteratorFunc: func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
				if *input.ShardId != "shard_1" {
					t.Errorf("expected shard id 'shard_1', got '%v':", *input.ShardId)
				}
				if *input.ShardIteratorType != "LATEST" {
					t.Errorf("expected shard iterator type of 'LATEST', got '%v':", *input.ShardIteratorType)
				}
				return &kinesis.GetShardIteratorOutput{
					ShardIterator: aws.String("shard_iterator_1"),
				}, nil
			},
			getRecordsFunc: func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
				if *input.ShardIterator != "shard_iterator_1" {
					t.Errorf("expected shard iterator 'shard_iterator_1', got '%v':", *input.ShardIterator)
				}
				return &kinesis.GetRecordsOutput{
					Records: []*kinesis.Record{
						{
							SequenceNumber: aws.String("sequence_1"),
							Data:           []byte(`{ "keys": ["sequence_1_1", "sequence_1_2", "sequence_1_3"], "ts": "2018-06-11T14:00:00.000Z" }`),
						},
					},
				}, nil
			},
			expectedRecordCount: 1,
			expectedRead:        true,
			expectedTo:          "sequence_1",
		},
		{
			name: "continue reading the stream from where we left off",
			getShardIteratorFunc: func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
				if *input.ShardId != "shard_1" {
					t.Errorf("expected shard id 'shard_1', got '%v':", *input.ShardId)
				}
				if *input.ShardIteratorType != "AFTER_SEQUENCE_NUMBER" {
					t.Errorf("expected shard iterator type of 'AFTER_SEQUENCE_NUMBER', got '%v':", *input.ShardIteratorType)
				}
				if *input.StartingSequenceNumber != "sequence_number" {
					t.Errorf("expected starting sequence number of 'sequence_number', got '%v':", *input.StartingSequenceNumber)
				}
				return &kinesis.GetShardIteratorOutput{
					ShardIterator: aws.String("shard_iterator_1"),
				}, nil
			},
			getRecordsFunc: func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
				if *input.ShardIterator != "shard_iterator_1" {
					t.Errorf("expected shard iterator 'shard_iterator_1', got '%v':", *input.ShardIterator)
				}
				return &kinesis.GetRecordsOutput{
					Records: []*kinesis.Record{
						{
							SequenceNumber: aws.String("sequence_1"),
							Data:           []byte(`{ "keys": ["sequence_1_1", "sequence_1_2", "sequence_1_3"], "ts": "2018-06-11T14:00:00.000Z" }`),
						},
					},
				}, nil
			},
			from:                "sequence_number",
			expectedRecordCount: 1,
			expectedRead:        true,
			expectedTo:          "sequence_1",
		},
	}

	for _, test := range tests {
		s := NewStream("test")
		s.svc = TestKinesisStream{
			GetShardIteratorFunc: test.getShardIteratorFunc,
			GetRecordsFunc:       test.getRecordsFunc,
		}
		records, to, read, err := s.getRecords("shard_1", test.from)
		if err != nil {
			t.Fatalf("%s: unexpected error getting records: %v", test.name, err)
		}
		if read != test.expectedRead {
			t.Errorf("%s: expected read to be %v, but was %v", test.name, test.expectedRead, read)
		}
		if len(records) != test.expectedRecordCount {
			t.Errorf("%s: expected to read %v records, but read %v", test.name, test.expectedRecordCount, len(records))
		}
		if to != test.expectedTo {
			t.Errorf("%s: expected to read until %v, but read until %v", test.name, test.expectedTo, to)
		}
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name                 string
		listShardsFunc       func(input *kinesis.ListShardsInput) (*kinesis.ListShardsOutput, error)
		getRecordsFunc       func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error)
		getShardIteratorFunc func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error)
		from                 StreamPosition
		expectedIDs          []string
		expectedTo           StreamPosition
	}{
		{
			name: "get single shard of records",
			listShardsFunc: func(input *kinesis.ListShardsInput) (op *kinesis.ListShardsOutput, err error) {
				op = &kinesis.ListShardsOutput{
					Shards: []*kinesis.Shard{
						{
							ShardId: aws.String("shard_1"),
						},
					},
				}
				return
			},
			getShardIteratorFunc: func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
				if *input.ShardId != "shard_1" {
					t.Errorf("expected shard id 'shard_1', got '%v':", *input.ShardId)
				}
				if *input.ShardIteratorType != "LATEST" {
					t.Errorf("expected shard iterator type of 'LATEST', got '%v':", *input.ShardIteratorType)
				}
				return &kinesis.GetShardIteratorOutput{
					ShardIterator: aws.String("shard_iterator_1"),
				}, nil
			},
			getRecordsFunc: func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
				if *input.ShardIterator != "shard_iterator_1" {
					t.Errorf("expected shard iterator 'shard_iterator_1', got '%v':", *input.ShardIterator)
				}
				return &kinesis.GetRecordsOutput{
					Records: []*kinesis.Record{
						{
							SequenceNumber: aws.String("sequence_1"),
							Data:           []byte(`{ "keys": ["sequence_1_1", "sequence_1_2", "sequence_1_3"], "ts": "2018-06-11T14:00:00.000Z" }`),
						},
					},
				}, nil
			},
			expectedTo: map[ShardID]SequenceNumber{
				ShardID("shard_1"): SequenceNumber("sequence_1"),
			},
			expectedIDs: []string{"sequence_1_1", "sequence_1_2", "sequence_1_3"},
		},
		{
			name: "get multiple shards, multiple records",
			listShardsFunc: func(input *kinesis.ListShardsInput) (op *kinesis.ListShardsOutput, err error) {
				if input.NextToken == nil {
					op = &kinesis.ListShardsOutput{
						NextToken: aws.String("next_shard"),
						Shards: []*kinesis.Shard{
							{
								ShardId: aws.String("shard_1"),
							},
						},
					}
					return
				}
				if *input.NextToken == "next_shard" {
					op = &kinesis.ListShardsOutput{
						Shards: []*kinesis.Shard{
							{
								ShardId: aws.String("shard_2"),
							},
						},
					}
					return
				}
				err = fmt.Errorf("listShardsFunc: unexpected next token: %v", input.NextToken)
				return
			},
			getShardIteratorFunc: func(input *kinesis.GetShardIteratorInput) (*kinesis.GetShardIteratorOutput, error) {
				if *input.ShardIteratorType != "LATEST" {
					t.Errorf("expected shard iterator type of 'LATEST', got '%v':", *input.ShardIteratorType)
				}
				if *input.ShardId == "shard_1" {
					return &kinesis.GetShardIteratorOutput{
						ShardIterator: aws.String("shard_iterator_1"),
					}, nil
				}
				if *input.ShardId == "shard_2" {
					return &kinesis.GetShardIteratorOutput{
						ShardIterator: aws.String("shard_iterator_2"),
					}, nil
				}
				return nil, fmt.Errorf("getShardIteratorFunc: unexpected shard id: %v", *input.ShardId)
			},
			getRecordsFunc: func(input *kinesis.GetRecordsInput) (*kinesis.GetRecordsOutput, error) {
				if *input.ShardIterator == "shard_iterator_1" {
					return &kinesis.GetRecordsOutput{
						Records: []*kinesis.Record{
							{
								SequenceNumber: aws.String("sequence_1"),
								Data:           []byte(`{ "keys": ["sequence_1_1", "sequence_1_2", "sequence_1_3"], "ts": "2018-06-11T14:00:00.000Z" }`),
							},
						},
					}, nil
				}
				if *input.ShardIterator == "shard_iterator_2" {
					return &kinesis.GetRecordsOutput{
						Records: []*kinesis.Record{
							{
								SequenceNumber: aws.String("sequence_2"),
								Data:           []byte(`{ "keys": ["sequence_2_1", "sequence_2_2", "sequence_2_3"], "ts": "2018-06-11T14:01:00.000Z" }`),
							},
							{
								SequenceNumber: aws.String("sequence_3"),
								Data:           []byte(`{ "keys": ["sequence_2_4", "sequence_2_5", "sequence_2_6"], "ts": "2018-06-11T14:01:00.000Z" }`),
							},
						},
					}, nil
				}
				return nil, fmt.Errorf("getRecordsFunc: unexpected shard iterator: %v", *input.ShardIterator)
			},
			expectedTo: map[ShardID]SequenceNumber{
				ShardID("shard_1"): SequenceNumber("sequence_1"),
				ShardID("shard_2"): SequenceNumber("sequence_3"),
			},
			expectedIDs: []string{"sequence_1_1", "sequence_1_2", "sequence_1_3", "sequence_2_1", "sequence_2_2", "sequence_2_3", "sequence_2_4", "sequence_2_5", "sequence_2_6"},
		},
	}

	for _, test := range tests {
		s := NewStream("test")
		s.svc = TestKinesisStream{
			ListShardsFunc:       test.listShardsFunc,
			GetShardIteratorFunc: test.getShardIteratorFunc,
			GetRecordsFunc:       test.getRecordsFunc,
		}
		ids, to, err := s.Get(test.from)
		if err != nil {
			t.Fatalf("%s: unexpected error getting records: %v", test.name, err)
		}
		if !reflect.DeepEqual(ids, test.expectedIDs) {
			t.Errorf("%s: expected IDs: '%v' but got '%v'", test.name, test.expectedIDs, ids)
		}
		if !reflect.DeepEqual(to, test.expectedTo) {
			t.Errorf("%s: expected to have final positions %v, but had %v", test.name, test.expectedTo, to)
		}
	}
}

func TestPut(t *testing.T) {
	var putRecordsCallCount, recordsAdded, itemsInRecords int
	tests := []struct {
		name                     string
		maxPutSize               int
		maxRecordSize            int
		putRecordsFunc           func(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
		ids                      []string
		expectedErr              error
		expectedPutRecordsCalled int
		expectedRecordsAdded     int
		expectedKeysInRecords    int
	}{
		{
			name:          "put records",
			maxRecordSize: 10, // bytes
			putRecordsFunc: func(input *kinesis.PutRecordsInput) (op *kinesis.PutRecordsOutput, err error) {
				putRecordsCallCount++
				recordsAdded += len(input.Records)
				for _, r := range input.Records {
					var data StreamData
					err := json.Unmarshal(r.Data, &data)
					if err != nil {
						t.Errorf("put records: unexpected error unmarhsalling data received by putRecords: %v", err)
						break
					}
					itemsInRecords += len(data.Keys)
				}
				return &kinesis.PutRecordsOutput{}, err
			},
			ids: []string{"12345", "67890", "12345"},
			expectedPutRecordsCalled: 1, // The call is below the default 2MB of data to PUT data, so only one call is expected.
			expectedRecordsAdded:     2, // Because the maxRecordSize is 10, { 12345, 67890 } should be in one record and { 123456 } in another.
			expectedKeysInRecords:    3, // 3 IDs should have been added to the stream.
		},
		{
			name:          "put record, need to make 2 calls due to limits on put size",
			maxPutSize:    10,
			maxRecordSize: 10, // bytes
			putRecordsFunc: func(input *kinesis.PutRecordsInput) (op *kinesis.PutRecordsOutput, err error) {
				putRecordsCallCount++
				recordsAdded += len(input.Records)
				for _, r := range input.Records {
					var data StreamData
					err := json.Unmarshal(r.Data, &data)
					if err != nil {
						t.Errorf("put records: unexpected error unmarhsalling data received by putRecords: %v", err)
						break
					}
					itemsInRecords += len(data.Keys)
				}
				return &kinesis.PutRecordsOutput{}, err
			},
			ids: []string{"12345", "67890", "12345"},
			expectedPutRecordsCalled: 2, // The call is below the default 2MB of data to PUT data, so only one call is expected.
			expectedRecordsAdded:     2, // Because the maxRecordSize is 10, { 12345, 67890 } should be in one record and { 123456 } in another.
			expectedKeysInRecords:    3, // 3 IDs should have been added to the stream.
		},
	}

	for _, test := range tests {
		putRecordsCallCount, recordsAdded, itemsInRecords = 0, 0, 0

		s := NewStream("test")
		s.svc = TestKinesisStream{
			PutRecordsFunc: test.putRecordsFunc,
		}
		if test.maxRecordSize > 0 {
			s.maxRecordSize = test.maxRecordSize
		}
		if test.maxPutSize > 0 {
			s.maxPutSize = test.maxPutSize
		}
		err := s.Put(test.ids)
		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("%s: expected error '%v', got '%v'", test.name, test.expectedErr, err)
		}
		if putRecordsCallCount != test.expectedPutRecordsCalled {
			t.Errorf("%s: expected putRecordsFunc to be called %d times, but was called %d times",
				test.name, test.expectedPutRecordsCalled, putRecordsCallCount)
		}
		if recordsAdded != test.expectedRecordsAdded {
			t.Errorf("%s: expected putRecordsFunc to be called %d times with %d records, but %d records were added",
				test.name, test.expectedPutRecordsCalled, test.expectedRecordsAdded, recordsAdded)
		}
		if itemsInRecords != test.expectedKeysInRecords {
			t.Errorf("%s: expected putRecordsFunc to be called %d times with %d records containing %d keys, but %d keys were added",
				test.name, test.expectedPutRecordsCalled, test.expectedRecordsAdded, test.expectedKeysInRecords, itemsInRecords)
		}
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		input    []string
		size     int
		expected [][]string
	}{
		{
			input:    []string{"A", "B", "C"},
			size:     1,
			expected: [][]string{{"A"}, {"B"}, {"C"}},
		},
		{
			input:    []string{"A", "B", "C"},
			size:     2,
			expected: [][]string{{"A", "B"}, {"C"}},
		},
		{
			input:    []string{"A", "B", "C"},
			size:     3,
			expected: [][]string{{"A", "B", "C"}},
		},
		{
			input:    []string{"A", "B", "C"},
			size:     200,
			expected: [][]string{{"A", "B", "C"}},
		},
		{
			input:    []string{"12345", "67890", "12345"},
			size:     10,
			expected: [][]string{{"12345", "67890"}, {"12345"}},
		},
	}

	for _, test := range tests {
		actual := chunk(test.input, test.size)
		if !reflect.DeepEqual(test.expected, actual) {
			t.Errorf("for input '%v' and chunk size %d, expected '%v', got '%v'", test.input, test.size,
				test.expected, actual)
		}
	}
}
