package snowflake

import (
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
	once sync.Once
)

// Init initializes the snowflake node with a given nodeID.
// It sets a custom epoch to "2024-01-01" to extend the lifespan of IDs.
func Init(nodeID int64) error {
	var err error
	once.Do(func() {
		// Set Epoch to 2024-01-01
		// This must be done before creating the node.
		// Note: snowflake.Epoch is int64 milliseconds.
		startTime, _ := time.Parse("2006-01-02", "2024-01-01")
		snowflake.Epoch = startTime.UnixNano() / 1000000

		node, err = snowflake.NewNode(nodeID)
	})
	if err != nil {
		return fmt.Errorf("failed to initialize snowflake node: %w", err)
	}
	return nil
}

// GenID generates a new unique uint64 ID.
// It panics if the node has not been initialized.
func GenID() uint64 {
	if node == nil {
		panic("snowflake node not initialized: call snowflake.Init() first")
	}
	return uint64(node.Generate().Int64())
}
