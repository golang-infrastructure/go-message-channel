package message_channel

// ------------------------------------------------ ---------------------------------------------------------------------

// CloseEventListener channel被关闭时的监听器
type CloseEventListener func()

// ------------------------------------------------ ---------------------------------------------------------------------

// ChannelConsumerFunc 用于消费channel中的元素
type ChannelConsumerFunc[Message any] func(index int, message Message)

// ------------------------------------------------ ---------------------------------------------------------------------

// ChannelOptions 创建Channel时的选项
type ChannelOptions[Message any] struct {

	// 关闭Channel时的回调函数
	CloseEventListener CloseEventListener

	// 用于消费channel中的元素
	ChannelConsumerFunc ChannelConsumerFunc[Message]

	// channel的缓存大小
	ChannelBuffSize uint64
}

// ------------------------------------------------ ---------------------------------------------------------------------
