package server

import (
	"database/sql"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

func addMessageRoutes(router *gin.RouterGroup, db *sql.DB) {
	message := router.Group("/message")
	{
		message.POST("/", func(ctx *gin.Context) {
			handleSaveMessage(ctx, db)
		})
		message.GET("/", func(c *gin.Context) {
			handleGetMessages(c, db)
		})
		message.PUT("/:id", func(c *gin.Context) {
			handleEditMessage(c, db)
		})
		message.DELETE("/:id", func(c *gin.Context) {
			handleDeleteMessage(c, db)
		})
	}
}

func handleDeleteMessage(c *gin.Context, db *sql.DB) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM Message WHERE ID = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "failed to delete message"})
		log.Fatal("failed to delete message")
	}

	c.JSON(200, gin.H{"success": true})
}

func handleSaveMessage(c *gin.Context, db *sql.DB) {
	message := Message{}
	err := c.BindJSON(&message)
	if err != nil {
		log.Fatal("failed to read message body")
		c.JSON(500, gin.H{"success": false, "error": "failed to read message body"})
	}

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"success": false, "error": "failed to start transaction"})
	}

	_, err = tx.Exec("INSERT INTO Message (ChatID, UserID, TextContent) VALUES ? , ? , ?", message.ChatID, message.UserID, message.TextContent)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		c.JSON(500, gin.H{"success": false, "error": "failed to save message"})
	}

	if len(message.Attachaments) > 0 {
		for _, attachament := range message.Attachaments {
			_, err = tx.Exec("INSERT INTO Attachament (MessageID, Type, Link) VALUES ? , ? , ?", attachament.MessageID, attachament.Type, attachament.Link)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				c.JSON(500, gin.H{"success": false, "error": "failed to save attachament"})
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"success": false, "error": "failed to commit transaction"})
	}

	c.JSON(200, gin.H{"success": true})
}

func handleGetMessages(c *gin.Context, db *sql.DB) {
	params := c.Request.URL.Query()
	chatID := params["chatID"][0]
	from := params["from"][0]
	limit := params["limit"][0]

	rows, err := db.Query("SELECT * FROM Message WHERE ChatID = ? ORDER BY TIMESTAMP DESC LIMIT ? OFFSET ?", chatID, limit, from)
	if err != nil {
		log.Fatal("failed to get messages")
		c.JSON(500, gin.H{"success": false, "error": "failed to get messages"})
	}

	messages := []Message{}
	for rows.Next() {
		message := Message{}
		err = rows.Scan(&message.ID, &message.ChatID, &message.UserID, &message.TextContent, &message.Timestamp, &message.ReplyToId, &message.WasEdited)
		if err != nil {
			log.Fatal("failed to read message")
			c.JSON(500, gin.H{"success": false, "error": "failed to read message"})
		}
		sem := make(chan struct{}, 4)
		wg := &sync.WaitGroup{}
		for _, message := range messages {
			wg.Add(1)
			sem <- struct{}{}
			go func(message *Message) {
				defer func() {
					wg.Done()
					<-sem
				}()
				attachments := []Attachament{}
				attachamentRows, err := db.Query("SELECT * FROM Attachament WHERE MessageID = ?", message.ID)
				if err != nil {
					log.Fatal("failed to get attachaments")
					c.JSON(500, gin.H{"success": false, "error": "failed to get attachaments"})
				}
				for attachamentRows.Next() {
					attachament := Attachament{}
					err = attachamentRows.Scan(&attachament.ID, &attachament.MessageID, &attachament.Type, &attachament.Link)
					if err != nil {
						log.Fatal("failed to read attachament")
						c.JSON(500, gin.H{"success": false, "error": "failed to read attachament"})
					}
					attachments = append(attachments, attachament)
				}
				message.Attachaments = attachments
			}(&message)
		}
		wg.Wait()
		close(sem)
	}

	c.JSON(200, gin.H{"success": true, "messages": messages})
}

func handleEditMessage(c *gin.Context, db *sql.DB) {
	// Get message ID and new message text from request body
	var reqBody struct {
		MessageID int    `json:"message_id"`
		Text      string `json:"text"`
	}
	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update message in database
	stmt, err := db.Prepare("UPDATE messages SET TextContent = ? WHERE ID = ?")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement"})
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(reqBody.Text, reqBody.MessageID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message updated successfully"})
}
