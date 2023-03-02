package message_channel

import (
	"sync"
	"sync/atomic"
	"time"
)

// 用于生成全局唯一ID
var idGenerator = atomic.Uint64{}

type MapRunFunc[Message any] func(m map[uint64]*Channel[Message])

type ChildrenMap[Message any] struct {
	lock       *sync.RWMutex
	channelMap map[uint64]*Channel[Message]
}

func NewChildrenMap[Message any]() *ChildrenMap[Message] {
	return &ChildrenMap[Message]{}
}

// Run 在map上执行函数，此函数的执行是串行互斥并发安全的
func (x *ChildrenMap[Message]) Run(f MapRunFunc[Message]) {
	x.lock.Lock()
	defer x.lock.Unlock()

	f(x.channelMap)
}

// BlockUtilEmpty 阻塞住直到当前map为空，期间会每隔给定的描述尝试获取map的情况
func (x *ChildrenMap[Message]) BlockUtilEmpty(f MapRunFunc[Message], interval ...time.Duration) {

	if len(interval) == 0 {
		interval = append(interval, time.Second*3)
	}

	isMapNotEmpty := false
	for {
		x.Run(func(m map[uint64]*Channel[Message]) {

			// 判断map是否为空了
			if len(m) == 0 {
				isMapNotEmpty = false
				return
			}

			// 调用用户处理map的函数
			if f != nil {
				f(m)
			}
		})

		// 判断是否需要结束
		if !isMapNotEmpty {
			break
		}

		// 休眠指定的时长
		time.Sleep(interval[0])
	}
}

// Size 统计map中元素的数量
func (x *ChildrenMap[Message]) Size() (size int) {
	x.Run(func(f map[uint64]*Channel[Message]) {
		size = len(f)
	})
	return
}

// ChildrenSlice 把map中所有的子channel都转为切片形式返回
func (x *ChildrenMap[Message]) ChildrenSlice() (childrenSlice []*Channel[Message]) {
	x.Run(func(f map[uint64]*Channel[Message]) {
		for _, channel := range f {
			childrenSlice = append(childrenSlice, channel)
		}
	})
	return
}

// Set 设置map的值
func (x *ChildrenMap[Message]) Set(id uint64, childChannel *Channel[Message]) (childrenSlice []*Channel[Message]) {
	x.Run(func(f map[uint64]*Channel[Message]) {
		f[id] = childChannel
	})
	return
}

// Remove 从map中删除值
func (x *ChildrenMap[Message]) Remove(id uint64) (childrenSlice []*Channel[Message]) {
	x.Run(func(f map[uint64]*Channel[Message]) {
		delete(f, id)
	})
	return
}
