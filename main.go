package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vmorsell/openai-gpt-sdk-go/gpt"
)

const (
	apiKey = ""

	userName        = "user"
	assistantName   = "CheapGPT"
	systemMessage   = "You are ChatGPT's cousin CheapGPT. You are just as good, but way less expensive."
	defaultChatName = "new-chat"
)

type Chat struct {
	Name     *string
	Messages []gpt.Message
	}

func main() {
	chat := Chat{}
	app := tview.NewApplication()

	chatView := tview.NewTextView()
	chatView.SetDynamicColors(true)

	infoBar := tview.NewTextView()
	infoBar.SetBackgroundColor(tcell.ColorBlue)
	fmt.Fprint(infoBar, "GPT-3.5")

	input := tview.NewInputField()
	input.
		SetLabel(fmt.Sprintf("%s ", fmtChatName(defaultChatName))).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack)

	sendLock := false

	updateCh := make(chan struct{})
	input.SetDoneFunc(func(_ tcell.Key) {
		if sendLock {
			return
		}

		message := input.GetText()
		if message == "" {
			return
		}

		sendLock = true

		fmt.Fprint(chatView, fmtChatMessage(fmt.Sprintf("[yellow]%s[white]", userName), message))
		input.SetText("")
		chat.Messages = append(chat.Messages, gpt.Message{
			Role:    gpt.RoleUser,
			Content: message,
		})
		updateCh <- struct{}{}
	})

	grid := tview.NewGrid().
		SetRows(0, 1, 1).
		AddItem(chatView, 0, 0, 1, 1, 0, 0, false).
		AddItem(infoBar, 1, 0, 1, 1, 0, 0, false).
		AddItem(input, 2, 0, 1, 1, 0, 0, true)

	client := gpt.NewClient(gpt.NewConfig().WithAPIKey(apiKey))
	fmt.Println(client)

	go func() {
		for {
			<-updateCh

			if chat.Name == nil {
				go func() {
					name, err := chatName(client, chat.Messages[len(chat.Messages)-1].Content)
					if err != nil {
						panic(fmt.Sprintf("chat name: %v", err))
					}
					chat.Name = &name
					input.SetLabel(fmt.Sprintf("%s ", fmtChatName(name)))
				}()
			}

			in := gpt.ChatCompletionInput{
				Model:    gpt.GPT35Turbo,
				Messages: chat.Messages,
				Stream:   true,
			}
			events := make(chan *gpt.ChatCompletionChunkEvent)
			go func() {
				err := client.ChatCompletionStream(in, events)
				if err != nil {
					panic(err)
				}
			}()

			fmt.Fprint(chatView, fmtChatMessage(assistantName, ""))
			app.Draw()

			for {
				event, ok := <-events
				if !ok {
					break
				}
				if event.Choices == nil {
					panic("no choices")
				}
				if event.Choices[0].Delta.Content == nil {
					// We may get events with no content. Eliminate this?
					continue
				}
				fmt.Fprintf(chatView, "%s", *event.Choices[0].Delta.Content)
				app.Draw()
			}
			sendLock = false
		}
	}()

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

func chatName(client gpt.Client, message string) (string, error) {
	res, err := client.ChatCompletion(gpt.ChatCompletionInput{
		Messages: []gpt.Message{
			{
				Role:    gpt.RoleSystem,
				Content: "You are a title generator. You generate names on the format foo-bar-baz. The titles have maximum three words, all lowercase, and the words are separated by dashes. No spaces are allowed. You always reply with only the word, nothing else.",
			},
			{
				Role:    gpt.RoleUser,
				Content: fmt.Sprintf("Can you suggest a title for a chat that started with this message?\n\n%s", message),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("chat completion: %w", err)
	}

	if res.Choices == nil {
		return "", fmt.Errorf("got 0 choices")
	}

	return res.Choices[0].Message.Content, nil
}

func fmtChatName(name string) string {
	return fmt.Sprintf("[#%s]", name)
}

func fmtChatMessage(user, message string) string {
	return fmt.Sprintf("\n%s <%s> %s", time.Now().Format("15:04"), user, message)
}
