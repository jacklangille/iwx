package messaging

import "fmt"

type Subjects struct {
	PlaceOrderBase string
	PartitionCount int
}

func (s Subjects) PlaceOrderForContract(contractID int64) string {
	return s.PlaceOrderForPartition(s.PlaceOrderPartition(contractID))
}

func (s Subjects) PlaceOrderForPartition(partition int) string {
	return fmt.Sprintf("%s.partition.%d", s.PlaceOrderBase, partition)
}

func (s Subjects) PlaceOrderForPartitionWildcard() string {
	return fmt.Sprintf("%s.partition.*", s.PlaceOrderBase)
}

func (s Subjects) PlaceOrderPartition(contractID int64) int {
	if s.PartitionCount <= 1 {
		return 0
	}

	value := int(contractID % int64(s.PartitionCount))
	if value < 0 {
		value += s.PartitionCount
	}

	return value
}
