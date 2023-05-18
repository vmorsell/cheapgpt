package main

import (
	"fmt"
	"time"

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
	Time    time.Time
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

	t := msg.Time.Format("15:04")
	line := fmt.Sprintf("%s <%s> %s\n", t, from, msg.Content)
	fmt.Fprint(c.ChatView, line)
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

			// Trigger the agent unless a trigger is already awaiting.
			select {
			case a.Trigger <- struct{}{}:
			default:
			}
		}
	}
}
