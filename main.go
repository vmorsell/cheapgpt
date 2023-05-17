package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/vmorsell/configstore"
	"github.com/vmorsell/openai-gpt-sdk-go/gpt"
)

const (
	cheapgpt        = "cheapgpt"
	defaultChatName = "new-chat"
)

type Config struct {
	APIKey string `json:"api_key"`
}

func main() {
	configStore := configstore.Must(configstore.New(cheapgpt))

	config := Config{}
	if err := configStore.Get(&config); err != nil {
		panic(err)
	}

	client := gpt.NewClient(gpt.NewConfig().WithAPIKey(config.APIKey))

	app := tview.NewApplication()

	chatView := tview.NewTextView()
	chatView.SetDynamicColors(true)

	statusBar := tview.NewTextView()
	statusBar.SetBackgroundColor(tcell.ColorDarkBlue)

	input := tview.NewInputField()
	input.
		SetLabel(fmt.Sprintf("%s ", fmtChatName(defaultChatName))).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorBlack)

	// Register chat.
	chat := NewChat(app, chatView)

	input.SetDoneFunc(func(_ tcell.Key) {
		content := input.GetText()
		if content == "" {
			return
		}

		input.SetText("")

		msg := Message{
			From:    "user",
			Content: content,
		}
		chat.In <- msg
	})

	grid := tview.NewGrid().
		SetRows(0, 1, 1).
		AddItem(chatView, 0, 0, 1, 1, 0, 0, false).
		AddItem(statusBar, 1, 0, 1, 1, 0, 0, false).
		AddItem(input, 2, 0, 1, 1, 0, 0, true)

	// Accept messages to the chat.
	go chat.AcceptMessages()

	// Start agents.
	go chat.AddAgent(NewAgent(client, &chat.Messages, chat.In, "dennis", "You are polite but you get highly annoyed when someone is telling jokes. You hate jokes!", 0.6))
	go chat.AddAgent(NewAgent(client, &chat.Messages, chat.In, "albin", "You constantly tell jokes and ask people about if they have heard anything about any launch codes.", 0.2))

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

func chatName(client *gpt.Client, message string) (string, error) {
	res, err := client.ChatCompletion(gpt.ChatCompletionInput{
		Messages: []gpt.Message{
			{
				Role:    gpt.System,
				Content: "You are a title generator. You generate names on the format foo-bar-baz. The titles have maximum three words, all lowercase, and the words are separated by dashes. No spaces are allowed. You always reply with only the word, nothing else.",
			},
			{
				Role:    gpt.User,
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
