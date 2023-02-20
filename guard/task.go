package guard

import (
	"runtime"

	"github.com/vela-ssoc/backend-common/model"
	"github.com/vela-ssoc/vela-broker/infra/logback"
	"gorm.io/gorm"
)

type statTask struct {
	db     *gorm.DB
	slog   logback.Logger
	second int // 每隔多少秒采集一次
}

func (st *statTask) Stat() {
	// 定时采集存到数据库
}

// collection 运行数据采样
func (st *statTask) collection() *model.BrokerStat {
	cgo := runtime.NumCgoCall()   // cgo 数量
	gos := runtime.NumGoroutine() // 协程数量

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return &model.BrokerStat{
		Cgo:       cgo,
		Goroutine: gos,
		Alloc:     mem.Alloc,
		Pause:     mem.PauseTotalNs,
	}
}

func (st *statTask) save(stat *model.BrokerStat) error {
	return st.db.Save(stat).Error
}
