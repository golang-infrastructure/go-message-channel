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

	// 每个channel都可以有一个名字
	Name string

	// 关闭Channel时的回调函数
	CloseEventListener CloseEventListener

	// 用于消费channel中的元素
	ChannelConsumerFunc ChannelConsumerFunc[Message]

	// channel的缓存大小
	ChannelBuffSize uint64
}

func NewChannelOptions[Message any]() *ChannelOptions[Message] {
	return &ChannelOptions[Message]{}
}

func (x *ChannelOptions[Message]) WithName(name string) *ChannelOptions[Message] {
	x.Name = name
	return x
}

func (x *ChannelOptions[Message]) WithCloseEventListener(closeEventListener CloseEventListener) *ChannelOptions[Message] {
	x.CloseEventListener = closeEventListener
	return x
}

func (x *ChannelOptions[Message]) WithChannelConsumerFunc(channelConsumerFunc ChannelConsumerFunc[Message]) *ChannelOptions[Message] {
	x.ChannelConsumerFunc = channelConsumerFunc
	return x
}

func (x *ChannelOptions[Message]) WithChannelBuffSize(channelBuffSize uint64) *ChannelOptions[Message] {
	x.ChannelBuffSize = channelBuffSize
	return x
}

// ------------------------------------------------ ---------------------------------------------------------------------
