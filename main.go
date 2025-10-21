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
	g.window.SetTitle("websocket-chat-gui v0.0.1")
	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("Host")
	roomIdEntry := widget.NewPasswordEntry() // Use NewPasswordEntry for masked input
	roomIdEntry.SetPlaceHolder("Room ID")
	roomPassEntry := widget.NewPasswordEntry() // Use NewPasswordEntry for masked input
	roomPassEntry.SetPlaceHolder("Room Password")
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")
	encryptionKeyEntry := widget.NewPasswordEntry() // Use NewPasswordEntry for masked input
	encryptionKeyEntry.SetPlaceHolder("Encryption Key")
	messageLabel := widget.NewLabel("")
	loginButton := widget.NewButton("Login", func() {
		g.client = &websocket_chat_client.Client{
			Conn:       nil,
			Addr:       addrEntry.Text,
			Proto:      "ws",
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
	testButton := widget.NewButton("TestLogin", func() {
		g.client = &websocket_chat_client.Client{
			Conn:       nil,
			Addr:       "10.0.0.10:8080",
			Proto:      "ws",
			RoomID:     "akkTrnwMkFKsqFf4",
			RoomPass:   "CEZv5XWFrgnifA3I",
			Username:   "GUI-TEST v0.0.1",
			MessageKey: "3ThCOI8DjsJ2G1O7",
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
		testButton,
	)
	g.window.SetContent(content)
}

func (g *GUI) chatWindow() {
	g.window.SetTitle("TODO : roomid")
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
				fyne.DoAndWait(func() {
					g.chatOutput.AppendMarkdown(fmt.Sprintf("%s: %s", rms.un, rms.msg))
					g.chatOutput.Refresh()
					g.scrollContainer.ScrollToBottom()
				})
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
					fyne.DoAndWait(func() {
						g.chatOutput.AppendMarkdown(fmt.Sprintf("User Count: %d", globalUserCount))
						g.chatOutput.Refresh()
						g.scrollContainer.ScrollToBottom()
					})
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
			fyne.DoAndWait(func() {
				g.chatOutput.AppendMarkdown(fmt.Sprintf("%s: %s", g.client.Username, nom))
				g.chatOutput.Refresh()
				g.scrollContainer.ScrollToBottom()
			})
		}
	}()
	return nil
}

func main() {
	g := GUI{}
	g.app = app.New()
	lifecycle(g)
	g.window = g.app.NewWindow("Login")
	g.window.Resize(fyne.NewSize(400, 700))
	g.loginWindow()
	g.window.ShowAndRun()
}
