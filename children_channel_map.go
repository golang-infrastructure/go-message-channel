package message_channel

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// 用于生成全局唯一ID，这样一次进程启动的生命周期中分配的ID都是唯一的
var idGenerator = atomic.Uint64{}

// MapRunFunc 用来处理map函数
// ctx: 用来做超时控制
// m: 传入的函数，可以并发安全的去操作
// error: 如果发生错误时会返回错误，返回nil为无错误
type MapRunFunc[Message any] func(ctx context.Context, m map[uint64]*Channel[Message]) error

// ChildrenMap 线程安全的map，用于存放当前信道的子信道
type ChildrenMap[Message any] struct {

	// 用于全局互斥操作
	lock *sync.Mutex

	// 用于记录当前channel的子channel
	channelMap map[uint64]*Channel[Message]
}

// NewChildrenMap 创建一个存放子channel的map
func NewChildrenMap[Message any]() *ChildrenMap[Message] {
	return &ChildrenMap[Message]{
		lock:       &sync.Mutex{},
		channelMap: make(map[uint64]*Channel[Message]),
	}
}

// Run 在map上执行函数，此函数的执行是串行互斥并发安全的
func (x *ChildrenMap[Message]) Run(ctx context.Context, f MapRunFunc[Message]) error {
	x.lock.Lock()
	defer x.lock.Unlock()

	return f(ctx, x.channelMap)
}

// BlockUtilEmpty 阻塞住直到当前map为空，期间会每隔给定的间隔尝试获取map的情况
func (x *ChildrenMap[Message]) BlockUtilEmpty(ctx context.Context, f MapRunFunc[Message], interval ...time.Duration) error {

	// 设置间隔的默认值，默认是每隔1秒试探一次
	if len(interval) == 0 {
		interval = append(interval, time.Second)
	}

	isMapNotEmpty := false
	for {
		err := x.Run(ctx, func(ctx context.Context, m map[uint64]*Channel[Message]) error {

			// 判断map是否为空了
			if len(m) == 0 {
				isMapNotEmpty = false
				return nil
			}

			// 调用用户处理map的函数
			if f != nil {
				return f(ctx, m)
			}

			return nil
		})
		if err != nil {
			return err
		}

		// 判断是否需要结束
		if !isMapNotEmpty {
			break
		}

		// 休眠指定的时长
		time.Sleep(interval[0])
	}

	return nil
}

// Size 统计map中元素的数量
func (x *ChildrenMap[Message]) Size(ctx context.Context) (int, error) {
	size := 0
	return size, x.Run(ctx, func(ctx context.Context, f map[uint64]*Channel[Message]) error {
		size = len(f)
		return nil
	})
}

// ChildrenSlice 把map中所有的子channel都转为切片形式返回
func (x *ChildrenMap[Message]) ChildrenSlice(ctx context.Context) ([]*Channel[Message], error) {
	childrenSlice := make([]*Channel[Message], 0)
	return childrenSlice, x.Run(ctx, func(ctx context.Context, f map[uint64]*Channel[Message]) error {
		for _, channel := range f {
			childrenSlice = append(childrenSlice, channel)
		}
		return nil
	})
}

// Set 设置map的值
func (x *ChildrenMap[Message]) Set(ctx context.Context, id uint64, childChannel *Channel[Message]) error {
	return x.Run(ctx, func(ctx context.Context, f map[uint64]*Channel[Message]) error {
		f[id] = childChannel
		return nil
	})
}

// Remove 从map中删除值
func (x *ChildrenMap[Message]) Remove(ctx context.Context, id uint64) error {
	return x.Run(ctx, func(ctx context.Context, f map[uint64]*Channel[Message]) error {
		delete(f, id)
		return nil
	})
}
