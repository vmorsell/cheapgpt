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

	userName      = "user"
	assistantName = "CheapGPT"
	systemMessage = "You are ChatGPT's cousin CheapGPT. You are just as good, but way less expensive."
)

func main() {
	messages := []gpt.Message{
		{
			Role:    gpt.RoleSystem,
			Content: systemMessage,
		},
	}

	app := tview.NewApplication()

	chat := tview.NewTextView()
	chat.SetBorder(true)

	input := tview.NewInputField()
	input.
		SetLabel("#cheapgpt ").
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

		fmt.Fprint(chat, fmtChatMessage(userName, message))
		input.SetText("")
		messages = append(messages, gpt.Message{
			Role:    gpt.RoleUser,
			Content: message,
		})
		updateCh <- struct{}{}
	})

	grid := tview.NewGrid().
		SetRows(0, 1).
		AddItem(chat, 0, 0, 1, 1, 0, 0, false).
		AddItem(input, 1, 0, 1, 1, 0, 0, true)

	client := gpt.NewClient(gpt.NewConfig().WithAPIKey(apiKey))
	fmt.Println(client)

	go func() {
		for {
			<-updateCh

			in := gpt.ChatCompletionInput{
				Model:    gpt.GPT35Turbo,
				Messages: messages,
				Stream:   true,
			}
			events := make(chan *gpt.ChatCompletionChunkEvent)
			go func() {
				err := client.ChatCompletionStream(in, events)
				if err != nil {
					panic(err)
				}
			}()

			fmt.Fprint(chat, fmtChatMessage(assistantName, ""))
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
				fmt.Fprintf(chat, "%s", *event.Choices[0].Delta.Content)
				app.Draw()
			}
			sendLock = false
		}
	}()

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

func fmtChatMessage(user, message string) string {
	return fmt.Sprintf("\n%s <%s> %s", time.Now().Format("15:04"), user, message)
}
