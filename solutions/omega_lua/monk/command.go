package monk

import (
	"fmt"
	"math/rand"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"

	"github.com/google/uuid"
)

type MonkCmdSender struct {
}

type CmdSender interface {
	SendWOCmd(cmd string)
	SendWSCmd(cmd string)
	SendPlayerCmd(cmd string)
	SendWSCmdAndInvokeOnResponse(string, func(output *packet.CommandOutput))
	SendPlayerCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
}

func NewMonkCmdSender() CmdSender {
	return &MonkCmdSender{}
}

func (c *MonkCmdSender) SendWSCmd(cmd string) {
	fmt.Printf("SendWSCmd: %s\n", cmd)
}

func (c *MonkCmdSender) SendWOCmd(cmd string) {
	fmt.Printf("SendWOCmd: %s\n", cmd)
}

func (c *MonkCmdSender) SendPlayerCmd(cmd string) {
	fmt.Printf("SendPlayerCmd: %s\n", cmd)
}

func (c *MonkCmdSender) genFakeCommandOutput() *packet.CommandOutput {
	return &packet.CommandOutput{
		OutputType:   packet.CommandOutputTypeDataSet,
		SuccessCount: uint32(rand.Intn(100)),
		OutputMessages: []protocol.CommandOutputMessage{
			protocol.CommandOutputMessage{
				Success:    true,
				Message:    "cmd resp hello: " + uuid.New().String(),
				Parameters: []string{"1", "2", "3"},
			},
			protocol.CommandOutputMessage{
				Success:    true,
				Message:    "cmd resp hello2: " + uuid.New().String(),
				Parameters: []string{"4", "5", "6"},
			},
		},
	}
}

func (c *MonkCmdSender) SendWSCmdAndInvokeOnResponse(cmd string, cb func(output *packet.CommandOutput)) {
	fmt.Printf("SendWSCmdAndInvokeOnResponse: %s\n", cmd)
	go func() {
		time.Sleep(time.Second)
		cb(c.genFakeCommandOutput())
	}()
}

func (c *MonkCmdSender) SendPlayerCmdAndInvokeOnResponseWithFeedback(cmd string, cb func(output *packet.CommandOutput)) {
	fmt.Printf("SendPlayerCmdAndInvokeOnResponseWithFeedback: %s\n", cmd)
	go func() {
		time.Sleep(time.Second)
		cb(c.genFakeCommandOutput())
	}()
}
