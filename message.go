package message_channel

import "sync"

// Channel 用于把多个Channel连接为一个channel，这样就可以用基于channel构建更复杂的通信模型
// 每个Channel对象是对go原生的channel的一个包装，同时增加了一些功能
type Channel[Message any] struct {

	// 全局唯一的ID，每个信道的ID都不同
	ID uint64

	// 真实存储数据的channel，每个channel都有一个消息发送方和消息接收方
	channel chan Message

	// 此信道创建的子信道，子信道会被连接到父信道，同时父信道要等所有的子信道退出后才能退出
	childrenChannelMap *ChildrenMap[Message]

	// 自己的负责处理消息的协程的退出标志位
	selfWorkerWg *sync.WaitGroup

	// 创建信道时的选项
	options *ChannelOptions[Message]
}

// NewChannel 创建一个信道
func NewChannel[Message any](options *ChannelOptions[Message]) *Channel[Message] {

	x := &Channel[Message]{
		ID:                 idGenerator.Add(1),
		channel:            make(chan Message, options.ChannelBuffSize),
		options:            options,
		childrenChannelMap: NewChildrenMap[Message](),
		selfWorkerWg:       &sync.WaitGroup{},
	}

	// 启动处理消息的协程
	x.selfWorkerWg.Add(1)
	go func() {

		defer func() {

			// 退出的时候需要设置自己的退出标记位
			x.selfWorkerWg.Done()

			// 同时退出的时候如果有事件回调的话需要触发一下事件回调
			if x.options.CloseEventListener != nil {
				x.options.CloseEventListener()
			}

		}()

		// 开始消费，处理channel
		count := 1
		for message := range x.channel {
			if x.options.ChannelConsumerFunc != nil {
				x.options.ChannelConsumerFunc(count, message)
			}
		}
	}()

	return x
}

// Send 往当前的消息队列中发送一条消息，消息会被异步处理
func (x *Channel[Message]) Send(message Message) {
	x.channel <- message
}

// MakeChildChannel 创建一条新的消息队列，对接到当前的消息队列上作为一个子队列
// 当前队列关闭之前需要等待所有的孩子队列关闭
func (x *Channel[Message]) MakeChildChannel() *Channel[Message] {

	subChannel := NewChannel[Message](&ChannelOptions[Message]{

		// 创建一个子信道，并将子信道上的所有消息都转发到父信道上
		ChannelConsumerFunc: func(index int, message Message) {
			x.channel <- message
		},

		// 子信道的缓存大小和父信道保持一致
		ChannelBuffSize: x.options.ChannelBuffSize,
	})

	// 在子信道关闭的时候告知父信道自己已经退出了
	subChannel.options.CloseEventListener = func() {
		x.childrenChannelMap.Remove(subChannel.ID)
	}

	// Adds a semaphore to the parent channel
	x.childrenChannelMap.Set(subChannel.ID, subChannel)

	return subChannel
}

// ReceiverWait 消息的接收方调用的，消息的接收方需要同步等待此消息信道被处理完毕时调用
func (x *Channel[Message]) ReceiverWait() {
	// 消息接收方等待发送消息的协程退出就认为是信道已经处理完了
	x.selfWorkerWg.Wait()
}

// SenderWaitAndClose 消息的发送方调用，消息的发送方需要同步等待消息被处理完时调用
func (x *Channel[Message]) SenderWaitAndClose(f ...MapRunFunc[Message]) {

	if len(f) == 0 {
		f = append(f, nil)
	}

	// 等待子channel消费完成退出
	x.childrenChannelMap.BlockUtilEmpty(f[0])

	// 关闭channel表示发送者不会再发送了，发送完队列中剩余的想这些就要退出了
	close(x.channel)

	// 等待消费完队列中剩余的消息
	x.selfWorkerWg.Wait()
}
