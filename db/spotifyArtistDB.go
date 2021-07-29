package Database

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	artistsSpotify []SpotifyArtists

	spotifyArtistPortalId int
	spotifyArtistPortal   string
	spotifyArtistTitle    string
)

type SpotifyArtists struct {
	id    int
	url   string
	title string
}

func SpotifyArtistConnectDB() {
	db, err := sql.Open("mysql", SpotifyDBUser+":"+SpotifyDBPass+"@tcp("+SpotifyDBHost+SpotifyDBPort+")/"+SpotifyDBName)
	if err != nil {
		log.Fatal("Cannot open remote DB connection", err)
	}

	// Spotify Artists
	rows, err := db.Query("select * from spotifyArtists")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&spotifyArtistPortalId, &spotifyArtistPortal, &spotifyArtistTitle); err != nil {
			panic(err)
		}

		artistSpotify := SpotifyArtists{spotifyArtistPortalId, spotifyArtistPortal, spotifyArtistTitle}
		artistsSpotify = append(artistsSpotify, artistSpotify) // add remote DB queries to a local reference
	}

	// Randomly select a artist to stream
	rand.Seed(time.Now().UnixNano())
	numArtists := len(artistsSpotify) // get number of accounts stored in remote DB
	min2 := 1
	max2 := numArtists
	artist := rand.Intn(max2 - min2 + 1)

	spotifyArtistPortal = artistsSpotify[artist].url
	spotifyArtistTitle = artistsSpotify[artist].title

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	db.Close()
}

func ReturnSpotifyArtistPortal() string {
	return spotifyArtistPortal
}

func ReturnSpotifyArtistTitle() string {
	return spotifyArtistTitle
}
