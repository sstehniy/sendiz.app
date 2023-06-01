package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func configureDB() *sql.DB {
	db := connectToDB()

	deleteTables(db)
	setupTables(db)
	query := `INSERT INTO User (FullName, Handle, Phone) VALUES ('Sergey Stehniy', 'sstehniy', '+380631234567')`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func connectToDB() *sql.DB {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func deleteTables(db *sql.DB) {
	_, err := db.Exec(`DROP TABLE IF EXISTS Attachament`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS Message`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS ChatMember`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS Chat`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS User`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS UserInitiate`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`DROP TABLE IF EXISTS UserVerification`)
	if err != nil {
		log.Fatal(err)
	}

}

func setupTables(db *sql.DB) {

	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS User (
		ID INT PRIMARY KEY AUTO_INCREMENT,
		FullName VARCHAR(255) NOT NULL,
		Handle VARCHAR(100) NOT NULL UNIQUE,
		Phone VARCHAR(15) NOT NULL UNIQUE,
		AvatarLink TEXT DEFAULT NULL
);
`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Chat (
			ID INT PRIMARY KEY AUTO_INCREMENT,
			Name VARCHAR(255) NOT NULL,
			ChatType VARCHAR(10) NOT NULL
		)`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS ChatMember (
			ID INT PRIMARY KEY AUTO_INCREMENT,
			ChatID INT NOT NULL,
			UserID INT NOT NULL,
			Role VARCHAR(10) NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Message (
			ID INT PRIMARY KEY AUTO_INCREMENT,
			ChatID INT NOT NULL,
			UserID INT NOT NULL,
			TextContent VARCHAR(10000) DEFAULT NULL,
			Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			WasEdited BOOLEAN DEFAULT FALSE,
			ReplyToId INT DEFAULT NULL
		)`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Attachament (
			ID INT PRIMARY KEY AUTO_INCREMENT,
			MessageID INT NOT NULL,
			Type VARCHAR(10) NOT NULL,
			Link VARCHAR(10000) NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS UserVerification (
			ID INT PRIMARY KEY AUTO_INCREMENT,
			Code VARCHAR(6) NOT NULL,
			Phone VARCHAR(15) NOT NULL,
			Status VARCHAR(10) NOT NULL,
			Created DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS UserInitiate (
			ID INT PRIMARY KEY AUTO_INCREMENT,
			Phone VARCHAR(15) NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}
}
