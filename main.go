package main

import (
	"encoding/json"
	"fmt"
	"log"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
	"github.com/keithmartin1982/websocket_chat_client"
)

var (
	sendMessage     chan interface{}
	globalUserCount int
	background      bool = false
)

type receivedMessage struct {
	un  string
	msg string
}

type GUI struct {
	app             fyne.App
	window          fyne.Window
	client          *websocket_chat_client.Client
	chatOutput      *widget.RichText
	scrollContainer *container.Scroll
}

func (g *GUI) loginWindow() {
	g.window.SetTitle("websocket-chat-gui")
	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("Host")
	roomIdEntry := widget.NewPasswordEntry()
	roomIdEntry.SetPlaceHolder("Room ID")
	roomPassEntry := widget.NewPasswordEntry()
	roomPassEntry.SetPlaceHolder("Room Password")
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	encryptionKeyEntry := widget.NewPasswordEntry()
	encryptionKeyEntry.SetPlaceHolder("Encryption Key")
	messageLabel := widget.NewLabel("")
	loginButton := widget.NewButton("Login", func() {
		g.client = &websocket_chat_client.Client{
			Conn:       nil,
			Addr:       addrEntry.Text,
			Proto:      "wss",
			SelfSigned: true,
			RoomID:     roomIdEntry.Text,
			RoomPass:   roomPassEntry.Text,
			Username:   usernameEntry.Text,
			MessageKey: encryptionKeyEntry.Text,
			MessageChan: make(chan struct {
				Type int
				Data []byte
			}),
		}
		if err := g.startClient(); err != nil {
			messageLabel.SetText("Login Not Successful!")
		} else {
			messageLabel.SetText("Login Successful!")
			g.chatWindow()
		}
	})
	content := container.NewVBox(
		widget.NewLabel("Please log in"),
		addrEntry,
		roomIdEntry,
		roomPassEntry,
		usernameEntry,
		encryptionKeyEntry,
		loginButton,
		messageLabel,
	)
	g.window.SetContent(content)
}

func (g *GUI) chatWindow() {
	g.window.SetTitle(g.client.RoomID)
	g.chatOutput = widget.NewRichText()
	g.chatOutput.Wrapping = fyne.TextWrapWord
	g.scrollContainer = container.NewVScroll(g.chatOutput)
	msgEntry := widget.NewEntry()
	msgEntry.OnSubmitted = func(s string) {
		if len(s) > 0 {
			sendMessage <- s
			msgEntry.SetText("")
		}
	}
	content := container.New(layout.NewBorderLayout(nil, msgEntry, nil, nil),
		g.scrollContainer,
		msgEntry,
	)
	g.window.SetContent(content)
}

func (g *GUI) startClient() error {
	if err := g.client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	// incoming messages
	go func() {
		for {
			nim := <-g.client.MessageChan
			nm := websocket_chat_client.Msg{}
			switch nim.Type {
			case websocket.TextMessage:
				if err := json.Unmarshal(nim.Data, &nm); err != nil {
					log.Printf("json unmarshal error: %v", err)
				}
				rms := receivedMessage{
					un:  nm.Username,
					msg: nm.Message,
				}
				g.appendText(rms.un, rms.msg)
			case websocket.BinaryMessage:
				ucm := struct {
					UC int `json:"cc"`
				}{}
				if err := json.Unmarshal(nim.Data, &ucm); err != nil {
					log.Printf("error: unmarshalling incomming message: %v", err)
				}
				// TODO : handle user count
				if globalUserCount != ucm.UC {
					globalUserCount = ucm.UC
					g.appendText("User Count", globalUserCount)
				}
			}
		}
	}()
	// outgoing messages
	sendMessage = make(chan interface{})
	go func() {
		for {
			nom := <-sendMessage
			if err := g.client.SendMsg(nom.(string)); err != nil {
				log.Printf("send message: %v", err)
				return
			}
			g.appendText(g.client.Username, nom)
		}
	}()
	return nil
}

func (g *GUI) appendText(prefix, content any) {
	if background {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   fmt.Sprintf("Msg from: %v", prefix),
			Content: fmt.Sprintf("%v", content),
		})
	}
	fyne.DoAndWait(func() {
		g.chatOutput.AppendMarkdown(fmt.Sprintf("%v: %v", prefix, content))
		g.chatOutput.AppendMarkdown("---")
		g.chatOutput.Refresh()
		g.scrollContainer.ScrollToBottom()
	})
}

func (g *GUI) lifecycle() {
	// App battery usage unrestricted is required for background websocket connection
	lifecycle := g.app.Lifecycle()
	lifecycle.SetOnExitedForeground(func() {
		background = true
	})
	lifecycle.SetOnEnteredForeground(func() {
		background = false
	})
}

func main() {
	g := GUI{}
	g.app = app.New()
	g.window = g.app.NewWindow("Login")
	platformDo(g)
	g.loginWindow()
	g.window.ShowAndRun()
}
