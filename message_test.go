package message_channel

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewChannel(t *testing.T) {
	options := NewChannelOptions[string]().WithChannelConsumerFunc(func(index int, message string) {
		t.Log(fmt.Sprintf("index = %d, message = %s", index, message))
	})
	channel := NewChannel[string](options)
	assert.NotNil(t, channel)
}

func TestChannel_MakeChildChannel(t *testing.T) {

}

func TestChannel_Single(t *testing.T) {
	options := NewChannelOptions[string]().WithChannelConsumerFunc(func(index int, message string) {
		t.Log(fmt.Sprintf("index = %d, message = %s", index, message))
	})
	channel := NewChannel[string](options)
	go func() {
		for i := 0; i < 10; i++ {
			channel.Send(fmt.Sprintf("message %d", i))
			time.Sleep(time.Second)
		}
		channel.SenderWaitAndClose()
	}()
	channel.ReceiverWait()
}

func TestChannel_Very_Complex(t *testing.T) {

}
