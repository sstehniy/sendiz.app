package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
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

		user.POST("/verify-otp", func(c *gin.Context) {
			verifyOTP(c, db)
		})
		user.Use(authMiddleWare)

		user.GET("/me", func(c *gin.Context) {
			getMe(c, db)
		})

		user.POST("/me", func(c *gin.Context) {
			createMe(c, db)
		})

		user.PUT("/me", func(c *gin.Context) {
			updateMe(c, db)
		})
	}
}

func authMiddleWare(c *gin.Context) {
	auth := c.Request.Header.Get("Authorization")
	if auth == "" {
		c.JSON(401, gin.H{"error": "Authorization header required"})
		c.Abort()
		return
	}
	// parse the token and get the user id from it
	// if the user id is not in the token, return an error
	// if the user id is in the token, add it to the context
	// so that the next handler can access it
	userId, err := parseToken(auth)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	c.Set("userId", userId)

	c.Next()
}

// The function `parseToken` takes in a string parameter called `authString`.
// It splits the string into two parts using the space character as a separator.
func parseToken(authString string) (string, error) {
	parsedData := strings.Split(authString, " ")

	// If the length of `parsedData` is not equal to 2, it returns an error message.
	if len(parsedData) != 2 {
		return "", fmt.Errorf("invalid token")
	}

	// If the first element of `parsedData` is not "Bearer", it returns an error message.
	if parsedData[0] != "Bearer" {
		return "", fmt.Errorf("invalid token")
	}

	// Here we get our secret key from the environment variables.
	hmacSampleSecret := []byte(os.Getenv("JWT_SECRET"))

	// We try to parse the second element of `parsedData` as JWT token, which is passed as an argument to `jwt.Parse` function.
	tk, err := jwt.Parse(strings.Split(authString, " ")[1], func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		// If the signing method is not HMAC, it returns an error message.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the HMAC sample secret as the key for verifying the token signature.
		return hmacSampleSecret, nil
	})

	// If the token is valid and claims can be extracted from the token,
	// it checks if the `userId` claim exists and returns that value as string.
	if claims, ok := tk.Claims.(jwt.MapClaims); ok && tk.Valid {
		if userId, ok := claims["userId"].(string); ok {
			return userId, nil
		} else {
			return "", fmt.Errorf("invalid token")
		}
	} else {
		// If there is an error, it prints the error and returns a nil value for the error.
		fmt.Println(err)
		return "", nil
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

func verifyOTP(c *gin.Context, db *sql.DB) {
	// parse the verification id from the request body
	userVerificationRequestData := UserVerificationClient{}
	err := c.BindJSON(&userVerificationRequestData)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
	}

	userVerification := UserVerification{}
	err = db.QueryRow(`SELECT * FROM UserVerification WHERE ID = ?`, userVerificationRequestData.ID).Scan(
		&userVerification.ID, &userVerification.Phone, &userVerification.Status, &userVerification.Created,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "no verification found"})
	}

	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(userVerification.Phone)
	params.SetCode(userVerificationRequestData.Code)

	t := twilio.NewRestClient()
	resp, err := t.VerifyV2.CreateVerificationCheck(os.Getenv("TWILIO_VERIFY_SID"), params)
	if err != nil {
		c.JSON(500, gin.H{"error": "error verifying the phone number"})
	}

	var success bool
	if status := resp.Status; status == nil {
		success = false
	} else {
		success = true
	}

	// create a new user in the database
	user := User{}
	user.Phone = userVerification.Phone

	// check if user exists
	userRow := db.QueryRow(`SELECT * FROM User WHERE Phone = ?`, user.Phone)

	var userId int64
	if userRow.Err() != nil {
		// user doesn't exist, create a new user initiate
		userInitiate := UserInitiate{}
		userInitiate.Phone = user.Phone

		dbResult, err := db.Exec(`INSERT INTO UserInitiate (Phone) VALUES (?)`, userInitiate.Phone)
		if err != nil {
			c.JSON(500, gin.H{"error": "error verifying user"})
		}

		id, err := dbResult.LastInsertId()
		if err != nil {
			c.JSON(500, gin.H{"error": "error verifying user"})
		}

		userId = id
	} else {
		// user exists, get the user id
		err = userRow.Scan(&user.ID, &user.FullName, &user.AvatarLink, &user.Phone)
		if err != nil {
			c.JSON(500, gin.H{"error": "error verifying user"})
		}

		userId = user.ID
	}

	// create a new token with secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
	})

	//send the token back to the client
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(500, gin.H{"error": "error verifying user"})
	}

	c.JSON(200, gin.H{"token": signedToken, "success": success})
}

func getMe(c *gin.Context, db *sql.DB) {
	userId := c.GetString("userId")

	userRow := db.QueryRow(`
		SELECT * FROM User
		WHERE ID = ?
		`, userId)

	if userRow.Err() != nil {
		userInitiateRow := db.QueryRow(`
		SELECT * FROM UserInitiate
		WHERE ID = ?
		`, userId)
		userInitiate := UserInitiate{}

		err := userInitiateRow.Scan(&userInitiate.ID, &userInitiate.Phone)
		if err != nil {
			c.JSON(500, gin.H{"error": "error getting user"})
			return
		}

		c.JSON(200, gin.H{"user": userInitiate, "initiate": true})
	}

	user := User{}

	err := userRow.Scan(&user.ID, &user.FullName, &user.AvatarLink, &user.Handle, &user.Phone)
	if err != nil {
		c.JSON(500, gin.H{"error": "error getting user"})
		return
	}

	c.JSON(200, gin.H{"user": user, "initiate": false})
}

func updateMe(c *gin.Context, db *sql.DB) {
	userId := c.GetString("userId")

	// check if user is an initiate
	userInitiateRow := db.QueryRow(`
				SELECT * FROM UserInitiate
				WHERE ID = ?
				`, userId)

	userInitiate := UserInitiate{}

	err := userInitiateRow.Scan(&userInitiate.ID, &userInitiate.Phone)

	if err == nil {
		c.JSON(400, gin.H{"error": "user is an initiate"})
		return
	}

	userUpdateData := User{}

	err = c.BindJSON(&userUpdateData)
	if err != nil {
		log.Printf("invalid request body: %s", err.Error())
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	userRow := db.QueryRow(`
				SELECT * FROM User
				WHERE ID = ?
				`, userId)

	user := User{}

	err = userRow.Scan(&user.ID, &user.FullName, &user.AvatarLink, &user.Handle, &user.Phone)

	if err != nil {
		log.Printf("error getting user data: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error getting user data"})
		return
	}

	if userUpdateData.FullName != "" {
		user.FullName = userUpdateData.FullName
	}

	if userUpdateData.AvatarLink != "" {
		user.AvatarLink = userUpdateData.AvatarLink
	}

	if userUpdateData.Handle != "" {
		user.Handle = userUpdateData.Handle
	}

	if userUpdateData.Phone != "" {
		user.Phone = userUpdateData.Phone
	}

	res, err := db.Exec(`
				UPDATE User
				SET FullName = ?, AvatarLink = ?, Handle = ?, Phone = ?
				WHERE ID = ?
		`, user.FullName, user.AvatarLink, user.Handle, user.Phone, user.ID)

	if err != nil {
		log.Printf("error updating user: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error updating user"})
		return
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil || rowsAffected == 0 {
		log.Printf("error updating user: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error updating user"})
		return
	}

	c.JSON(200, gin.H{"success": true, "user": user})
}

func createMe(c *gin.Context, db *sql.DB) {
	userId := c.GetString("userId")
	type UserInput struct {
		Phone      string `json:"phone"`
		FullName   string `json:"fullName"`
		AvatarLink string `json:"avatarLink"`
		Handle     string `json:"handle"`
	}
	//read request body and save to user struct
	userData := UserInput{}
	err := c.BindJSON(&userData)

	if err != nil {
		log.Printf("invalid request body: %s", err.Error())
		c.JSON(400, gin.H{"success": false, "error": "invalid request body"})
		return
	}

	userInitiateRow := db.QueryRow(`
				SELECT * FROM UserInitiate
				WHERE ID = ?
				`, userId)

	if userInitiateRow.Err() != nil {
		log.Printf("user initiate not found: %s", userInitiateRow.Err().Error())
		c.JSON(400, gin.H{"success": false, "error": "user initiate is not found"})
		return
	}

	dbResult, err := db.Exec(`
				INSERT INTO User (Phone, FullName, AvatarLink, Handle)
				VALUES (?, ?, ?, ?)
		`, userData.Phone, userData.FullName, userData.AvatarLink, userData.Handle)

	if err != nil {
		log.Printf("error creating user: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error creating user"})
		return
	}

	rowsEffected, err := dbResult.RowsAffected()

	if err != nil || rowsEffected == 0 {
		log.Printf("error creating user: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error creating user"})
		return
	}
	id, err := dbResult.LastInsertId()

	if err != nil || rowsEffected == 0 {
		log.Printf("error creating user: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error creating user"})
		return
	}

	// delete user initiate
	_, err = db.Exec(`
				DELETE FROM UserInitiate
				WHERE ID = ?
		`, userId)

	if err != nil {
		log.Printf("error deleting user initiate: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error creating user"})
		return
	}

	userRow := db.QueryRow(`
				SELECT * FROM User
				WHERE ID = ?
		`, id)

	user := User{}

	err = userRow.Scan(&user.ID, &user.FullName, &user.AvatarLink, &user.Handle, &user.Phone)

	if err != nil {
		log.Printf("error reading user data: %s", err.Error())
		c.JSON(500, gin.H{"success": false, "error": "error creating user"})
		return
	}

	c.JSON(200, gin.H{"success": true, "user": user})
}
