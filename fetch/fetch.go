// Cycled imports workaround
package Fetch

import Database "moonlandr/db"

var (
	SpotifyId       int
	SpotifyEmail    string
	SpotifyPassword string
)

func Creds() {
	SpotifyId = Database.ReturnSpotifyId()
	SpotifyEmail = Database.ReturnSpotifyEmail()
	SpotifyPassword = Database.ReturnSpotifyPassword()
}
