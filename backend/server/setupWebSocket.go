package server

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func setupWebSocket(router *gin.Engine, db *sql.DB) {
	router.GET("/ws", wsHandler)
}

func wsHandler(c *gin.Context) {
	// Upgrade HTTP request to WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Listen indefinitely for new messages coming through on the WebSocket
	for {
		// Read message from browser
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		// Print the message to the console
		log.Printf("%s sent: %s\n", ws.RemoteAddr(), string(msg))

		// Echo the message back to the browser
		err = ws.WriteMessage(msgType, msg)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
