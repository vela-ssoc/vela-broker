package model

// BrokerStat 节点采集上报的数据
type BrokerStat struct {
	ID        int64  `json:"id"`        // 节点 ID
	Timestamp int64  `json:"timestamp"` // 秒级时间戳
	Cgo       int64  `json:"cgo"`       // cgo 数量
	Goroutine int    `json:"goroutine"` // 协程个数
	Alloc     uint64 `json:"alloc"`     // 占用内存
	Pause     uint64 `json:"pause"`     // GC 所用的总纳秒数
}

func (BrokerStat) TableName() string {
	return "broker_stat"
}
