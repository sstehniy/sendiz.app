package server

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

func addUserRoutes(router *gin.RouterGroup, db *sql.DB) {
	user := router.Group("/user")
	{
		user.POST("/send-otp", func(c *gin.Context) {
			sendOTP(c, db)
		})

	}
}

func sendOTP(c *gin.Context, db *sql.DB) {
	phone := c.Request.URL.Query().Get("phone")
	if phone == "" {
		c.JSON(400, gin.H{"error": "phone is required"})
		return
	}

	valid, err := validatePhoneNumber(phone)
	if err != nil || !valid {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	t := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})

	params := &verify.CreateVerificationParams{}
	params.SetTo(phone)
	params.SetChannel("sms")

	resp, err := t.VerifyV2.CreateVerification(os.Getenv("TWILIO_VERIFY_SID"), params)
	status := resp.Status
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
	} else {
		if status != nil {
			fmt.Println(*status)
			c.JSON(200, gin.H{"status": *status})
		} else {
			fmt.Println(status)
			c.JSON(200, gin.H{"success": "verification code sent"})
		}
	}

	userVerification := UserVerification{}
	userVerification.Phone = phone
	userVerification.Status = *status

	dbResult, err := db.Exec(`
		INSERT INTO UserVerification (Phone, Status)
		VALUES (?, ?)
	`, userVerification.Phone, userVerification.Status)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id, err := dbResult.LastInsertId()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"id": id})
}
