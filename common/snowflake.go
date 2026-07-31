package common

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	sfNode *snowflake.Node
	sfOnce sync.Once
)

// InitSnowflake 初始化雪花 ID 节点（机器ID 1-1023）
func InitSnowflake(nodeID int64) error {
	var err error
	sfOnce.Do(func() {
		sfNode, err = snowflake.NewNode(nodeID)
	})
	return err
}

// GenerateOrderNo 生成全局唯一订单号
func GenerateOrderNo() (string, error) {
	if sfNode == nil {
		return "", fmt.Errorf("snowflake not initialized")
	}
	return sfNode.Generate().String(), nil
}

// GenerateID 生成雪花 uint64 主键（门店 store_id 等）
func GenerateID() (uint64, error) {
	if sfNode == nil {
		return 0, fmt.Errorf("snowflake not initialized")
	}
	return uint64(sfNode.Generate().Int64()), nil
}
