package server

import "github.com/gorilla/websocket"

type User struct {
	ID         int64  `json:"id"`
	FullName   string `json:"fullName"`
	Handle     string `json:"handle"`
	Phone      string `json:"phone"`
	AvatarLink string `json:"avatarLink"`
}

type UserInitiate struct {
	ID    int64  `json:"id"`
	Phone string `json:"phone"`
}

type UserVerification struct {
	ID      int64  `json:"id"`
	Phone   string `json:"phone"`
	Status  string `json:"status"`
	Created string `json:"created"`
}

type UserVerificationClient struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
}

type ChatMember struct {
	ID     int64  `json:"id"`
	ChatID int64  `json:"chatId"`
	UserID int64  `json:"userId"`
	Role   string `json:"role"`
}

type Chat struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ChatType string `json:"chatType"`
	Members  []User `json:"members"`
}

type Attachament struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"messageId"`
	Type      string `json:"type"`
	Link      string `json:"link"`
}

type Message struct {
	ID           int64         `json:"id"`
	ChatID       int64         `json:"chatId"`
	UserID       int64         `json:"userId"`
	TextContent  string        `json:"content"`
	Attachaments []Attachament `json:"attachaments"`
	Timestamp    string        `json:"timestamp"`
	WasEdited    bool          `json:"wasEdited"`
	ReplyToId    int64         `json:"replyTo"`
}

type Client struct {
	id     int64
	socket *websocket.Conn
	send   chan []byte
}

type Hub struct {
	id         int64
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
}
