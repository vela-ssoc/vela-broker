package mlink

import "sync"

func container() subsection {
	// 128 * 64 = 8192
	const size = 128
	buckets := make([]*bucket, size)
	for i := 0; i < size; i++ {
		buckets[i] = &bucket{elems: make(map[int64]*connect, 64)}
	}

	return subsection{
		size:    size,
		buckets: buckets,
	}
}

// subsection 读写安全的分段 map
type subsection struct {
	size    int64     // 容量
	buckets []*bucket // 存储桶
}

// get 获取元素
func (sec *subsection) get(key int64) (any, bool) {
	bkt := sec.bucket(key)
	return bkt.get(key)
}

// put 存放并返回是否存放成功，如果 key 已经存在则存放失败
func (sec *subsection) put(key int64, val *connect) bool {
	bkt := sec.bucket(key)
	return bkt.put(key, val)
}

// del 删除元素并返回是否存在且删除成功
func (sec *subsection) del(key int64) bool {
	bkt := sec.bucket(key)
	return bkt.del(key)
}

// bucket 根据 key 计算所在的存储桶
func (sec *subsection) bucket(key int64) *bucket {
	idx := key % sec.size
	return sec.buckets[idx]
}

// foreach 循环遍历所有连接
func (sec *subsection) foreach(fn func(key int64, val *connect)) {
	defer func() { recover() }()
	for i := 0; i < int(sec.size); i++ {
		bkt := sec.buckets[i]
		elems := bkt.copy()
		for k, v := range elems {
			fn(k, v)
		}
	}
}

type bucket struct {
	mutex sync.RWMutex
	elems map[int64]*connect
}

func (bkt *bucket) get(key int64) (any, bool) {
	bkt.mutex.RLock()
	val, exist := bkt.elems[key]
	bkt.mutex.RUnlock()

	return val, exist
}

func (bkt *bucket) put(key int64, val *connect) bool {
	bkt.mutex.Lock()
	_, exist := bkt.elems[key]
	if !exist {
		bkt.elems[key] = val
	}
	bkt.mutex.Unlock()

	return !exist
}

func (bkt *bucket) del(key int64) bool {
	bkt.mutex.Lock()
	_, exist := bkt.elems[key]
	if exist {
		delete(bkt.elems, key)
	}
	bkt.mutex.Unlock()

	return exist
}

// copy 复制元素
func (bkt *bucket) copy() map[int64]*connect {
	bkt.mutex.RLock()
	ret := make(map[int64]*connect, len(bkt.elems))
	for k, v := range bkt.elems {
		ret[k] = v
	}
	bkt.mutex.RUnlock()

	return ret
}
