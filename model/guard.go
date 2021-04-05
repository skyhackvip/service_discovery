package model

import (
	"github.com/skyhackvip/service_discovery/configs"
	"sync"
	"sync/atomic"
)

type Guard struct {
	renewCount     int64
	lastRenewCount int64
	needRenewCount int64
	threshold      int64
	lock           sync.RWMutex
}

func (gd *Guard) incrNeed() {
	gd.lock.Lock()
	defer gd.lock.Unlock()
	gd.needRenewCount += int64(configs.CheckEvictInterval / configs.RenewInterval)
	gd.threshold = int64(float64(gd.needRenewCount) * configs.SelfProtectThreshold)
}

func (gd *Guard) decrNeed() {
	gd.lock.Lock()
	defer gd.lock.Unlock()
	gd.needRenewCount -= int64(configs.CheckEvictInterval / configs.RenewInterval)
	gd.threshold = int64(float64(gd.needRenewCount) * configs.SelfProtectThreshold)
}

func (gd *Guard) setNeed(count int64) {
	gd.lock.Lock()
	defer gd.lock.Unlock()
	gd.needRenewCount = count * int64(configs.CheckEvictInterval/configs.RenewInterval)
	gd.threshold = int64(float64(gd.needRenewCount) * configs.SelfProtectThreshold)
}

func (gd *Guard) incrCount() {
	atomic.AddInt64(&gd.renewCount, 1)
}

func (gd *Guard) storeLastCount() {
	atomic.StoreInt64(&gd.lastRenewCount, atomic.SwapInt64(&gd.needRenewCount, 0))
}

func (gd *Guard) selfProtectStatus() bool {
	return atomic.LoadInt64(&gd.lastRenewCount) < atomic.LoadInt64(&gd.threshold)
}
