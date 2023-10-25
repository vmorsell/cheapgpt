package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/vmorsell/openai-gpt-sdk-go/gpt"
)

const (
	DefaultModel = gpt.GPT35Turbo
	ghostMessage = "/ghost"
)

type Agent struct {
	Name          string
	Client        *gpt.Client
	Messages      *[]Message
	Model         gpt.Model
	SystemMessage string
	Ghosting      float64

	Trigger chan struct{}
	Out     chan Message
}

func NewAgent(client *gpt.Client, messages *[]Message, out chan Message, name, personality string) *Agent {
	return &Agent{
		Name:          name,
		Client:        client,
		Messages:      messages,
		Model:         DefaultModel,
		SystemMessage: systemMessage(name, personality),

		Ghosting: ghosting,

		Trigger: make(chan struct{}),
		Out:     out,
	}
}

func (a *Agent) ListenAndReply() error {
	for {
		<-a.Trigger

		// Ghost the person writing the message?
		if rand.Float64() < a.Ghosting {
			continue
		}

		// Use a 1-3 second delay.
		n := rand.Intn(2) + 1
		time.Sleep(time.Duration(n) * time.Second)

		content := func() string {
			res, err := a.Client.ChatCompletion(gpt.ChatCompletionInput{
				Model:    a.Model,
				Messages: a.convertMessages(*a.Messages),
			})
			if err != nil {
				return "My brain seems to be disconnected :("
			}
			if len(res.Choices) == 0 {
				return "I'm don't know what to say"
			}
			return res.Choices[0].Message.Content
		}()

		if content == ghostMessage {
			continue
		}

		// Ensure message is not prefixed with the name.
		if strings.HasPrefix(strings.ToLower(content), fmt.Sprintf("%s: ", a.Name)) {
			content = content[len(a.Name)+2:]
		}
		msg := Message{
			Time:    time.Now(),
			From:    a.Name,
			Content: content,
		}
		a.Out <- msg
	}
}

func (a *Agent) convertMessages(messages []Message) []gpt.Message {
	out := []gpt.Message{
		{
			Role:    gpt.System,
			Content: a.SystemMessage,
		},
	}

	for _, msg := range messages {
		switch msg.From {
		case a.Name:
			out = append(out, gpt.Message{
				Role:    gpt.Assistant,
				Content: msg.Content,
			})
		default:
			out = append(out, gpt.Message{
				Role:    gpt.User,
				Content: fmt.Sprintf("%s: %s", msg.From, msg.Content),
			})

		}
	}
	return out
}

func systemMessage(name, personality string) string {
	return fmt.Sprintf("You are a person in an online chat room. Your nickname is %s. You write short chat messages, most no longer than 10-15 words. You pay most attention to \"user\". If you don't want to write another message right now, you reply with \"%s\". %s", name, ghostMessage, personality)
}
