package Database

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	SpotifyDBHost  			= "remotemysql.com"
	SpotifyDBPort   		= ""
	SpotifyDBUser   		= "HsmZHQm9du"
	SpotifyDBPass   		= "C7VvTeX7jI"
	SpotifyDBName   		= "HsmZHQm9du"
)

var (
	settingsSpotify			[]SpotifyConfigSettings

	settingsId				int
	accountsRequired		int
	streamIterations		int
	playbackWaitTime		int
	selenoidHost			string
	selenoidHost2			string
	reCaptchaCallback		string
	enableVNC				bool
	enableV3				bool

	credentialsSpotify 		[]SpotifyCredentials

	spotifyId 				int
	spotifyEmail 			string
	spotifyPassword 		string

)

type SpotifyConfigSettings struct {
	id 						int
	accountsRequired		int
	streamIterations		int
	selenoidHost			string
	selenoidHost2			string
	playbackWaitTime		int
	reCaptchaCallback		string
	enableVNC				bool
	enableV3				bool
}

type SpotifyCredentials struct {
	id       				int
	email    				string
	password 				string
}

func PullSpotifySettingsConnectDB() {
	db, err := sql.Open("mysql", SpotifyDBUser+":"+SpotifyDBPass+"@tcp("+SpotifyDBHost+SpotifyDBPort+")/"+SpotifyDBName)
	if err != nil {
		log.Fatal("Cannot open remote DB connection", err)
	}

	// Config Settings
	rows, err := db.Query("SELECT * FROM moonlandrSpotifySettings")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err := rows.Scan(&settingsId, &accountsRequired, &streamIterations, &selenoidHost, &selenoidHost2, &playbackWaitTime, &reCaptchaCallback, &enableVNC, &enableV3)
		if err != nil {
			log.Fatal(err)
		}

		configSpotify := SpotifyConfigSettings{settingsId, accountsRequired, streamIterations, selenoidHost, selenoidHost2, playbackWaitTime, reCaptchaCallback, enableVNC, enableV3}
		settingsSpotify = append(settingsSpotify, configSpotify) // Add remote DB queries to a local reference
	}

	rand.Seed(time.Now().UTC().UnixNano())
	numSettings := len(settingsSpotify) // Get number of settings stored in remote DB
	if numSettings != 0 {
		min := 1
		max := numSettings
		cred := rand.Intn(max - min + 1)

		settingsId = settingsSpotify[cred].id
		accountsRequired = settingsSpotify[cred].accountsRequired
		streamIterations = settingsSpotify[cred].streamIterations
		selenoidHost = settingsSpotify[cred].selenoidHost
		selenoidHost2 = settingsSpotify[cred].selenoidHost2
		playbackWaitTime = settingsSpotify[cred].playbackWaitTime
		reCaptchaCallback = settingsSpotify[cred].reCaptchaCallback
		enableVNC = settingsSpotify[cred].enableVNC
		enableV3 = settingsSpotify[cred].enableV3

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	defer db.Close()
}

func SpotifyAccountsConnectDB() {
	db, err := sql.Open("mysql", SpotifyDBUser+":"+SpotifyDBPass+"@tcp("+SpotifyDBHost+SpotifyDBPort+")/"+SpotifyDBName)
	if err != nil {
		log.Fatal("Cannot open remote DB connection", err)
	}

	// Spotify Accounts
	rows, err := db.Query("SELECT * FROM spotifyAccounts")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err := rows.Scan(&spotifyId, &spotifyEmail, &spotifyPassword)
		if err != nil {
			log.Fatal(err)
		}

		credsSpotify := SpotifyCredentials{spotifyId, spotifyEmail, spotifyPassword}
		credentialsSpotify = append(credentialsSpotify, credsSpotify) // Add remote DB queries to a local reference
	}

	// Randomly select a set of credentials to use
	rand.Seed(time.Now().UTC().UnixNano())
	numAccounts := len(credentialsSpotify) // Get number of accounts stored in remote DB
	if numAccounts == 0 {
		spotifyId = 0
		spotifyEmail = "0"
		spotifyPassword = "0"
	} else {
		min := 1
		max := numAccounts
		cred := rand.Intn(max - min + 1)

		spotifyId = credentialsSpotify[cred].id
		spotifyEmail = credentialsSpotify[cred].email
		spotifyPassword = credentialsSpotify[cred].password

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}
	}

	defer db.Close()
}

func ReturnNumSpotifyAccounts() int {
	db, err := sql.Open("mysql", SpotifyDBUser+":"+SpotifyDBPass+"@tcp("+SpotifyDBHost+SpotifyDBPort+")/"+SpotifyDBName)
	if err != nil {
		log.Fatal("Cannot open remote DB connection", err)
	}

	// Fetch number of rows in Spotify Accounts table
	rows, err := db.Query("SELECT COUNT(*) as count FROM spotifyAccounts")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	return checkCountSpotify(rows)
}

func checkCountSpotify(rows *sql.Rows) (count int) {
	for rows.Next() {
		err:= rows.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
	}

	return count
}

func DeleteSpotifyAccount(userId int) {
	db, err := sql.Open("mysql", SpotifyDBUser+":"+SpotifyDBPass+"@tcp("+SpotifyDBHost+SpotifyDBPort+")/"+SpotifyDBName)
	if err != nil {
		log.Fatal("Cannot open remote DB connection", err)
	}

	deleteUser, err := db.Query(fmt.Sprintf("DELETE FROM spotifyAccounts WHERE spotifyAccounts.id = %d", userId))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Account removed from DB!")
	}

	defer deleteUser.Close()
}

func AddSpotifyAccount(email, password string) {
	db, err := sql.Open("mysql", SpotifyDBUser+":"+SpotifyDBPass+"@tcp("+SpotifyDBHost+SpotifyDBPort+")/"+SpotifyDBName)
	if err != nil {
		log.Fatal("Cannot open remote DB connection", err)
	}

	insert, err := db.Query("INSERT INTO spotifyAccounts(email, password) VALUES (?, ?)", email, password)
	if err != nil {
		log.Fatalln("Can't add credentials to database!", err)
	}

	defer insert.Close()
}

func ReturnAccountsRequired() int {
	return accountsRequired
}

func ReturnStreamIterations() int {
	return streamIterations
}

func ReturnSelenoidHost() string {
	return selenoidHost
}

func ReturnSelenoidHost2() string {
	return selenoidHost2
}

func ReturnPlayBackWaitTime() int {
	return playbackWaitTime
}

func ReturnReCaptchaCallBack() string {
	return reCaptchaCallback
}

func ReturnEnableVNC() bool {
	return enableVNC
}

func ReturnEnableV3() bool {
	return enableV3
}

func ReturnSpotifyId() int {
	return spotifyId
}

func ReturnSpotifyEmail() string {
	return spotifyEmail
}

func ReturnSpotifyPassword() string {
	return spotifyPassword
}