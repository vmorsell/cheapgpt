package main

import (
	"fmt"

	"github.com/rivo/tview"
)

type Chat struct {
	App      *tview.Application
	ChatView *tview.TextView
	Messages []Message
	Agents   []*Agent
	In       chan Message
}

type Message struct {
	From    string
	Content string
}

func NewChat(app *tview.Application, chatView *tview.TextView) *Chat {
	return &Chat{
		App:      app,
		ChatView: chatView,
		In:       make(chan Message),
	}
}

func (c *Chat) AddAgent(agent *Agent) {
	c.Agents = append(c.Agents, agent)
	go agent.ListenAndReply()
}

func (c *Chat) PrintMessage(msg Message) {
	from := msg.From
	if from == "user" {
		from = fmt.Sprintf("[yellow]%s[white]", from)
	}

	fmt.Fprint(c.ChatView, fmtChatMessage(from, msg.Content))
	c.App.Draw()
}

func (c *Chat) AcceptMessages() {
	for {
		msg := <-c.In
		c.Messages = append(c.Messages, msg)
		c.PrintMessage(msg)

		for _, a := range c.Agents {
			// Don't re-trigger the agent who sent the message.
			if a.Name == msg.From {
				continue
			}

			a.Trigger <- struct{}{}
		}
	}
}
