package main

type User struct {
	ID         int64  `json:"id"`
	FullName   string `json:"fullName"`
	Handle     string `json:"handle"`
	Status     string `json:"status"`
	Phone      string `json:"phone"`
	AvatarLink string `json:"avatarLink"`
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
