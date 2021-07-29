package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	Worker "moonlandr/worker"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

var (
	cyan 			= color.New(color.FgHiCyan, color.Bold)
	yellow 			= color.New(color.FgYellow, color.Bold)
	red 			= color.New(color.FgRed, color.Bold)
	magenta 		= color.New(color.FgMagenta, color.Bold)
	white			= color.New(color.FgWhite, color.Bold)

	workers			= [1]string{"Spotify"}
)

func main() {
	var service string

	// Initialize backbone CLI
	app := cli.NewApp()
	app.Name = "Moonlandr"
	app.Description = "It's more than just a Record Label."
	app.Version = "2.0.0"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "deploy",
			Destination: &service,
			Usage: "Deploy a Supported Service Worker. (For Ex: Spotify)",
		},
	}

	app.Action = func(c *cli.Context) error {
		appAcii := `
                    ▄▄▄█
                ▄▄▀█▒░▒
             ▄▐█▄ ░▒▒▒▒▒▄
            ▄█▒▒▒▒▒▀▒▒▒▒▒░▒▄
           ▀ ▀▓▄▒▒▒▒▄▄ ▀░▒▒▒▒
         █ ▄▀▒ ░░▒▒▒▒▒   ▒░▀
        ▐      ▐▒▒▓▒▒▒
                ░ ░▀▒░
             ▒▒▀  ▒▓▓
        ▄░░▒▒░▒▄░░█▄░▄▄
     ░▒▒▒▒▒░▒▒▒░ ▐░▒█░
   █░▒▒▒▒▒▓███▒▄▄▓░▒░ ▐▒
  █▌▒▒░░▒▒░█▒█████▌▒░░░▒▒
 ██▒▒▒░▒▒▒░░▒█▓█▓▀░▒▒▒▒▒▒▒
 ██░▓░▒▒▒▒▒▒▒▒▄▒▒▒▒░░▒▒░▒▒
 ███▒▒▒▒▒▒▒▒▒▒▒▒░▒▒█▒█▒▌▒░
 ▐███▒░▓▒░▒▒▒▒▒░█░▓██▒█░▒▌
  ▀███░▓██▒▒▒▒▒▒▒░▀▓▀░▒░▀
    ▀████░░▒░▒▒▒▒▒▒▒▒░▒
      ▀▀███▓▓▄▄▄▄▄▓▀▀
             ▀
                $$\      $$\                               $$\                           $$\           
                $$$\    $$$ |                              $$ |                          $$ |          
                $$$$\  $$$$ | $$$$$$\   $$$$$$\  $$$$$$$\  $$ | $$$$$$\  $$$$$$$\   $$$$$$$ | $$$$$$\  
                $$\$$\$$ $$ |$$  __$$\ $$  __$$\ $$  __$$\ $$ | \____$$\ $$  __$$\ $$  __$$ |$$  __$$\ 
                $$ \$$$  $$ |$$ /  $$ |$$ /  $$ |$$ |  $$ |$$ | $$$$$$$ |$$ |  $$ |$$ /  $$ |$$ |  \__|
                $$ |\$  /$$ |$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |$$  __$$ |$$ |  $$ |$$ |  $$ |$$ |      
                $$ | \_/ $$ |\$$$$$$  |\$$$$$$  |$$ |  $$ |$$ |\$$$$$$$ |$$ |  $$ |\$$$$$$$ |$$ |      
                \__|     \__| \______/  \______/ \__|  \__|\__| \_______|\__|  \__| \_______|\__|
		`

		_, _ = cyan.Printf("\n" + appAcii + "\n\tv" + app.Version + "\n\n")

		_, _ = white.Println("\t~ ] " + app.Description + "\n")

		_, _ = yellow.Printf("\tPROFITABLE WORKERS\n\n")
		_, _ = magenta.Printf("\t1. Spotify \n")

		// Flag parsing
		if service == "spotify" {
			Worker.SpotifyInit()
		} else {
			autoWorkerDeployer()
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Automatically Deploy a Service Worker
func autoWorkerDeployer() {
	platformAmount := len(workers) + 1 // # of Supported Platforms

	rand.Seed(time.Now().UTC().UnixNano())
	min := 1
	max := platformAmount
	selectedWorkerID := rand.Intn(max - min) + min

	if selectedWorkerID > 0 && selectedWorkerID <= platformAmount {
		_, _ = yellow.Printf("\n\tWorker #%d is deploying...\n", selectedWorkerID)

		time.Sleep(3 * time.Second) // wait
	}

	switch selectedWorkerID {
		case 1:
			Worker.SpotifyInit()
			break
		default:
			_, _ = red.Printf("\n\t! ] NO WORKER SELECTED!\n\n")

			Worker.SpotifyInit()
	}
}