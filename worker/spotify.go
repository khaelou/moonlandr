package Worker

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	Database "moonlandr/db"
	Fetch "moonlandr/fetch"
	Captcha "moonlandr/twocaptchaV3"

	"github.com/fatih/color"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var (
	green 					= color.New(color.FgHiGreen, color.Bold)
	cyan 					= color.New(color.FgHiCyan, color.Bold)
	yellow 					= color.New(color.FgYellow, color.Bold)
	red 					= color.New(color.FgRed, color.Bold)
	magenta 				= color.New(color.FgMagenta, color.Bold)

	accountsRequired			int
	streamIterations			int
	playbackWaitTime			int
	host					string
	captchaCallback				string
	enableVNC				bool
	enableV3				bool

	UserIdSpotify				int
	UserEmailSpotify 			string
	UserPasswordSpotify 			string
	UserGenEmailSpotifyGlbl			string
	UserGenPasswordSpotifyGlbl		string

	ArtistProfileSpotify			string
	ArtistNameSpotify			string
	mlPlaybackTime				string

	accountsInDB				= 0
	accountCreationCount			= accountsInDB + 1
	accountCreationErrorCount		= 0
	misplacedTrackCheckFail     		= 0
	shuffleIterationSpotify			= 1
	shuffleInvalidations			= 0
	estimatedFigCount			= 0.00000

	// Simply modify attributes here if Spotify where to make HTML changes
	incorrectLoginAlert			= "p[class='alert alert-warning']"
	artistNameLabel				= "h1[@class='_77ccab85bb794646035d39a339c41781-scss']"
	followBtn				= "button[class='b49b68b14a1771a4cb36313f2b350e84-scss']"
	unfollowBtn				= "button[class='b49b68b14a1771a4cb36313f2b350e84-scss _2b37b3630aee3cbfc97689b5af341d60-scss']"
	disabledSkipBtn				= "button[class='control-button spoticon-skip-forward-16 control-button--disabled']"
	playBtn					= "button[class='control-button spoticon-play-16 control-button--circled']"
	loadingPlayBtn				= "button[class='control-button spoticon-play-16 control-button--circled control-button--loading']"
	pauseBtn				= "button[class='control-button spoticon-pause-16 control-button--circled']"
	skipBtn					= "button[class='control-button spoticon-skip-forward-16']"
	playbackTimeDiv				= "div[class='playback-bar__progress-time e80fc2e59729be32410c909c47ef87a3-scss']"
	trackTitleDiv				= "div[class='c319b99793755cc3bba709fe1b1fda42-scss ellipsis-one-line']"
	artistNameDiv				= "div[class='_44843c8513baccb36b3fa171573a128f-scss ellipsis-one-line']"
	accMenuBtn				= "button[class='_34098cfd13d48e2910679f35aea2c377-scss']"
	logoutBtn				= "button[class='_5d8857b271ece35ed4dd191b5b15f48e-scss']"
	robotErrorLabel				= "label[class='has-error']"
	recaptchaError				= "div[class='FormHelpText__Help-e48exm-0 doEKrx InputErrorMessage__Container-tliowl-0 ciTMoJ']"
	yourLibraryNavBtn			= "div[class='navBar-link-text-with-icon-wrapper']"

	twoCaptchaAPIKey			= "06b1f801f4b0bcc0d1abea45e7306543" // 2captcha.com API key

	v2ReCaptchaKey				= "6Lenb9oUAAAAAO1Rqrm4KsoNH14OvMm6SWkQcdRj" // Pulled from 'SpotifyRegistrationURL'
	v3ReCaptchaKey				= "6LfDteEUAAAAAFW-ygeu6HCIkLwEM7HiV_Zl5Hu3" // Pulled from 'SpotifyRegistrationURL'
)

const (
	port            			= 4444 // 4444 (Selenoid Port)
	SpotifyRegistrationURL			= "https://www.spotify.com/us/signup/?forward_url=https%3A%2F%2Fopen.spotify.com%2F" // For account creation
)

func initArtistSpotify() {
	Database.SpotifyArtistConnectDB() // Open remote DB connection for artists
	ArtistProfileSpotify = Database.ReturnSpotifyArtistPortal()
	ArtistNameSpotify = Database.ReturnSpotifyArtistTitle()

	fmt.Printf("\n\tSTREAMING SPOTIFY ARTIST: " + ArtistNameSpotify + "\n")

	Database.SpotifyAccountsConnectDB() // Open remote DB connection for accounts
	Fetch.Creds() // Retrieve selected Spotify credentials for the session
	UserIdSpotify = Fetch.SpotifyId
	UserEmailSpotify = Fetch.SpotifyEmail
	UserPasswordSpotify = Fetch.SpotifyPassword
}

func restartSpotify(wd selenium.WebDriver) {
	wd.Quit() // Quit previous session
	SpotifyInit()
}

func hostBalancer() {
	host1 := Database.ReturnSelenoidHost() // Amsterdam
	host2 := Database.ReturnSelenoidHost2() // Frankfurt

	selectedHost := []string{
		host1,
		host2,
	}

	rand.Seed(time.Now().UnixNano())

	host = selectedHost[rand.Intn(len(selectedHost))]

	switch host {
	case host1:
		fmt.Println("\n\t— Amsterdam, Netherlands Selenoid Host.")
		break
	case host2:
		fmt.Println("\n\t— Frankfurt, Germany Selenoid Host.")
		break
	}
}

func SpotifyInit() {
	Database.PullSpotifySettingsConnectDB() // Open remote DB connection for config settings
	accountsInDB = Database.ReturnNumSpotifyAccounts() // Placed here to refresh value if more accounts are added
	accountsRequired = Database.ReturnAccountsRequired()
	streamIterations = Database.ReturnStreamIterations()
	playbackWaitTime = Database.ReturnPlayBackWaitTime()
	captchaCallback  = Database.ReturnReCaptchaCallBack()
	enableVNC = Database.ReturnEnableVNC()
	enableV3 = Database.ReturnEnableV3()

	hostBalancer() // Used to select a Selenoid Host, better manage CPU

	// Check if enough accounts are in remote DB
	if accountsInDB >= accountsRequired {
		fmt.Printf(fmt.Sprintf("\n\t* ] SPOTIFY WORKER: %d Accounts in DB Queue\n", accountsInDB))

		initArtistSpotify() // Pull credentials from remote DB
		loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify) // Authenticate Spotify account
	} else {
		caps := selenium.Capabilities{"browserName": "chrome", "enableVNC": enableVNC}
		chromeCaps := chrome.Capabilities{
			Path:  "",
			Args: []string{
				"--headless", // Runs chrome without any visible UI
				"--no-sandbox",
			},
		}
		caps.AddChrome(chromeCaps)
		wd, err := selenium.NewRemote(caps, fmt.Sprintf("%s:%d/wd/hub", host, port))
		if err != nil {
			panic(err)
		}
		defer wd.Quit()

		initSpotifyAccountCreation(wd)
	}
}

func loginProcessSpotify(id int, email, password string) {
	fmt.Printf(fmt.Sprintf("\n\tACC ID: %d", id))
	fmt.Printf("\n\tACCOUNT EMAIL: " + email)
	fmt.Printf("\n\tACCOUNT PASSWORD: " + password + "\n")
	fmt.Println()

	// Connect to the WebDriver instance running locally. (Selenoid)
	caps := selenium.Capabilities{"browserName": "chrome", "enableVNC": enableVNC}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("%s:%d/wd/hub", host, port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Additional check verifying if the Remote DB has any accounts stored, if not create some
	if id == 0 && email == "0" && password == "0" {
		wd.Quit()
		initSpotifyAccountCreation(wd)
	}

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(loginProcessSpotify) panic occured: ", r)

			wd.Refresh()
			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}()

	// Navigate to the (Login Page -> Artist Profile)
	if err := wd.Get(ArtistProfileSpotify); err != nil {
		panic(err)
	} else {
		_, _ = yellow.Println("\tPage reached.")

		if title, err := wd.Title(); err == nil {
			fmt.Printf("\n\t%s\n\n", title)
		} else {
			_,_ = red.Printf("\n\tFailed to get page title: %s\n", err)
			return
		}
	}

	// Email
	emailElem, err := wd.FindElement(selenium.ByID, "login-username")
	if err != nil {
		panic(err)
	}
	err = emailElem.SendKeys(UserEmailSpotify)
	if err != nil {
		log.Println("Email wasn't entered.")
	} else {
		log.Println("Email entered.")
	}

	// Password
	pwElem, err := wd.FindElement(selenium.ByID, "login-password")
	if err != nil {
		panic(err)
	}
	err = pwElem.SendKeys(UserPasswordSpotify)
	if err != nil {
		log.Println("Password wasn't entered.")
	} else {
		log.Println("Password entered.")
	}

	// Click login button
	btn, err := wd.FindElement(selenium.ByID, "login-button")
	if err != nil {
		panic(err)
	}
	if err := btn.Click(); err != nil {
		_,_ = red.Println("\n\tLogin button not clicked.\n")
	} else {
		log.Println("Login button clicked.")

		credentialCheckSpotify(wd)
	}
}

func credentialCheckSpotify(wd selenium.WebDriver) {
	time.Sleep(5 * time.Second) // Wait

	_, err := wd.FindElement(selenium.ByCSSSelector, incorrectLoginAlert)
	if err != nil {
		// Double check credentials
		_, err := wd.FindElement(selenium.ByXPATH, artistNameLabel)
		if err != nil {
			_,_ = cyan.Println("\n\tUser successfully authenticated!\n")

			time.Sleep(4 * time.Second) // Wait

			// Loading Play button check
			_, err := wd.FindElement(selenium.ByCSSSelector, loadingPlayBtn)
			if err != nil {
				// Unfollow button check
				_, err := wd.FindElement(selenium.ByCSSSelector, unfollowBtn)
				if err != nil {
					followArtistSpotify(wd)
					startDiscographySpotify(wd)
				} else {
					startDiscographySpotify(wd)
				}
			} else {
				panic(err)
			}
		} else {
			_,_ = red.Println("\n\tUser not authenticated, restart...\n")

			restartSpotify(wd)
		}
	} else {
		_, _ = red.Printf("\n\tInvalid credentials, empty DB table..\n")

		Database.DeleteSpotifyAccount(UserIdSpotify) // Remove old credentials from DB

		// Check if enough accounts are in remote DB
		if accountsInDB >= accountsRequired {
			restartSpotify(wd)
		} else {
			if accountsInDB == 0 {
				initSpotifyAccountCreation(wd)
			} else {
				restartSpotify(wd)
			}
		}
	}
}

func followArtistSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(followArtistSpotify) Follow button couldn't be found.")

			logoutSpotify(wd)

			_ = r // Do something with error
		}
	}()

	time.Sleep(4 * time.Second) // Wait

	followBtn, err := wd.FindElement(selenium.ByCSSSelector, followBtn)
	if err != nil {
		panic(err)
	}
	if err := followBtn.Click(); err != nil {

		panic(err)
	} else {
		_, _ = green.Println("\tNow following artist!\n")
	}
}

func startDiscographySpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); nil != r {
			fmt.Println("(startDiscographySpotify) Account currently in-use by another client.")

			logoutSpotify(wd)

			_ = r
		}
	}()

	time.Sleep(7 * time.Second) // Wait

	// Pause button check
	_, err := wd.FindElement(selenium.ByCSSSelector, pauseBtn)
	if err != nil {
		startBtn, err := wd.FindElement(selenium.ByCSSSelector, playBtn)
		if err != nil {
			panic(err)
		}
		if err := startBtn.Click(); err != nil {
			_,_ = red.Println("\n\tDiscography click failed.\n")
		} else {
			artistNameDiv, err := wd.FindElement(selenium.ByCSSSelector, artistNameDiv)
			if err != nil {
				panic(err)
			} else {
				artistName, _ := artistNameDiv.Text()

				if artistName != ArtistNameSpotify {
					fmt.Printf("\tSTREAMING SPOTIFY ARTIST/S: " + artistName + " [Correction]\n\n")
				}
			}

			log.Printf("(#%d) Artist Discography Started!", shuffleIterationSpotify)

			SpotifyRepeatShuffleEvery(time.Duration(playbackWaitTime) * time.Second, wd)
		}
	} else {
		pauseBtn, err := wd.FindElement(selenium.ByCSSSelector, pauseBtn)
		if err != nil {
			panic(err)
		}
		if err := pauseBtn.Click(); err != nil {
			_,_ = red.Println("\n\tPause button click failed.\n")
		} else {
			startDiscographySpotify(wd)
		}
	}
}

func shuffleSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			time.Sleep(time.Duration(playbackWaitTime) * time.Second)

			shuffleSpotify(wd)
		}
	}()

	// Advertisement check
	_, err := wd.FindElement(selenium.ByCSSSelector, disabledSkipBtn)
	if err != nil {
		playbackTimeDiv, err := wd.FindElement(selenium.ByCSSSelector, playbackTimeDiv)
		if err != nil {
			logoutSpotify(wd)
		}

		playbackTime, _ := playbackTimeDiv.Text()
		minuteFromTime :=  playbackTime[0:1] // removes ":00" from playback time
		minute, _ := strconv.ParseInt(minuteFromTime, 10, 64)
		secondsFromTime := playbackTime[2:len(playbackTime)] // removes "0:" from playback time
		seconds, _ := strconv.ParseInt(secondsFromTime, 10, 64)

		if seconds >= 10 {
			mlPlaybackTime = strconv.Itoa(int(minute)) + ":" + strconv.Itoa(int(seconds))
		} else {
			mlPlaybackTime = strconv.Itoa(int(minute)) + ":0" + strconv.Itoa(int(seconds))
		}

		if playbackTime == mlPlaybackTime {
			mlPlaybackTime = playbackTime
		} else {
			_,_ = red.Println("\n\tPlayback time error.\n")
		}

		trackTitleDiv, err := wd.FindElement(selenium.ByCSSSelector, trackTitleDiv)
		if err != nil {
			logoutSpotify(wd)
		}
		trackTitle, _ := trackTitleDiv.Text()

		// Shuffle button
		shuffleBtn, err := wd.FindElement(selenium.ByCSSSelector, skipBtn)
		if err != nil {
			panic(err)
		}
		if err := shuffleBtn.Click(); err != nil {
			_,_ = red.Println("\n\tShuffle button click failed.\n")
		} else {
			if seconds > 0 {
				// Check if playback surpassed 30 seconds
				if seconds >= 30 && minute <= 1 || seconds < 30 && minute > 0 {
					log.Printf("[✓] Track '%s' playback stopped at %s. [~$%s]", trackTitle, mlPlaybackTime, pricePerStreamGen(0.00331, 0.00437))

					estimatedRoyaltyCalc()

					// Shuffle/Skip button
					skipBtn, err := wd.FindElement(selenium.ByCSSSelector, skipBtn)
					if err != nil {
						panic(err)
					}

					if err := skipBtn.Click(); err != nil {
						shuffleInvalidations++
						_, _ = red.Printf("\tShuffle to next song failed. #(%d)\n", shuffleInvalidations)

						if shuffleInvalidations % 3 == 0 {
							shuffleInvalidations = 0

							wd.Refresh()
							startDiscographySpotify(wd)
						}
					} else {
						shuffleIterationSpotify++

						log.Printf("#(%d) Shuffle to next song.", shuffleIterationSpotify)

						misplacedTrackCheck(wd) // Check if there is a misplaced track before continuing
					}
				}
			}
		}
	} else {
		log.Println("Advertisement displayed, wait..")

		logoutSpotify(wd) // Restart
	}

	// Iteration check
	if shuffleIterationSpotify % streamIterations == 0 {
		_, _ = red.Println("\n\tStream limit reached, rotating..\n")

		logoutSpotify(wd)
	}
}

func SpotifyRepeatShuffleEvery(d time.Duration, wd selenium.WebDriver) {
	for _ = range time.Tick(d) {
		shuffleSpotify(wd)
	}
}

func logoutSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(logoutSpotify) panic occured: ", r)

			restartSpotify(wd)
		}
	}()

	_, err := wd.FindElement(selenium.ByCSSSelector, accMenuBtn)
	if err != nil {
		log.Println("Account menu not visible, overriding logout process.")

		panic(err)
	} else {
		menuBtn, err := wd.FindElement(selenium.ByCSSSelector, accMenuBtn)
		if err != nil {
			panic(err)
		}
		if err := menuBtn.Click(); err != nil {
			log.Println("Account menu click failed.")
		} else {
			logoutBtn, err := wd.FindElement(selenium.ByCSSSelector, logoutBtn)
			if err != nil {
				panic(err)
			}

			if err := logoutBtn.Click(); err != nil {
				_,_ = red.Println("\n\tLogout button click failed.\n")
			} else {
				_, _ = yellow.Println("\n\tSuccessfully logged out.")

				restartSpotify(wd)
			}
		}
	}
}

func initSpotifyAccountCreation(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(initSpotifyAccountCreation) panic occured: ", r)

			wd.Quit() // Close previous session
			initSpotifyAccountCreation(wd) // Restart
		}
	}()

	// Check if enough accounts are in remote DB
	if accountsInDB >= accountsRequired {
		restartSpotify(wd)
	} else {
		fmt.Printf(fmt.Sprintf("\n\t+ ] SPOTIFY ACCOUNT CREATION: %d Accounts in DB Queue\n", accountsInDB))

		rand.Seed(time.Now().UTC().UnixNano())

		min := 1
		max := 99999999
		randomVal := rand.Intn(max - min) + min

		UserGenEmailSpotify := fmt.Sprintf("%s%d%s", "landrmoon", randomVal, "@protonmail.com")
		UserGenPasswordSpotify := "SpaceCadet1231!"
		UserGenNameSpotify := fmt.Sprintf("%s%d", "betalandr", randomVal)

		UserGenEmailSpotifyGlbl = UserGenEmailSpotify
		UserGenPasswordSpotifyGlbl = UserGenPasswordSpotify

		// Connect to the WebDriver instance running. (Selenoid)
		caps := selenium.Capabilities{"browserName": "chrome", "enableVNC": enableVNC}
		chromeCaps := chrome.Capabilities{
			Path:  "",
			Args: []string{
				"--headless", // Runs chrome without any visible UI
				"--no-sandbox",
			},
		}
		caps.AddChrome(chromeCaps)
		wd, err := selenium.NewRemote(caps, fmt.Sprintf("%s:%d/wd/hub", host, port))
		if err != nil {
			panic(err)
		}
		defer wd.Quit()

		// Navigate to the (Registration Page)
		if err := wd.Get(SpotifyRegistrationURL); err != nil {
			panic(err)
		} else {
			_, _ = yellow.Println("\n\tPage reached.")

			if title, err := wd.Title(); err == nil {
				fmt.Printf("\n\t%s\n\n", title)
			} else {
				_,_ = red.Printf("\n\tFailed to get page title: %s\n", err)
				return
			}
		}

		v3Solver(wd, enableV3)

		// Email
		emailElem, err := wd.FindElement(selenium.ByID, "email")
		if err != nil {
			panic(err)
		}
		if err := emailElem.Clear(); err != nil {
			log.Println("Email wasn't cleared.")
		} else {
			log.Println("Email was cleared.")
		}
		err = emailElem.SendKeys(UserGenEmailSpotify)
		if err != nil {
			log.Println("Email wasn't entered.")
		} else {
			log.Println("Email entered. (" + UserGenEmailSpotify + ")")
		}

		// Confirm Email
		emailConfirmElem, err := wd.FindElement(selenium.ByID, "confirm")
		if err != nil {
			panic(err)
		}
		if err := emailConfirmElem.Clear(); err != nil {
			log.Println("Confirm Email wasn't cleared.")
		} else {
			log.Println("Confirm Email was cleared.")
		}
		err = emailConfirmElem.SendKeys(UserGenEmailSpotify)
		if err != nil {
			log.Println("Confirm Email wasn't entered.")
		} else {
			log.Println("Confirm Email entered.")
		}

		// Password
		pwElem, err := wd.FindElement(selenium.ByID, "password")
		if err != nil {
			panic(err)
		}
		err = pwElem.SendKeys(UserGenPasswordSpotify)
		if err != nil {
			log.Println("Password wasn't entered.")
		} else {
			log.Println("Password entered. (" + UserGenPasswordSpotify + ")")
		}

		// Display Name
		nameElem, err := wd.FindElement(selenium.ByID, "displayname")
		if err != nil {
			panic(err)
		}
		err = nameElem.SendKeys(UserGenNameSpotify)
		if err != nil {
			log.Println("Display Name wasn't entered.")
		} else {
			log.Println("Display Name entered.")
		}

		// DOB Selection
		monthElem, err := wd.FindElement(selenium.ByID, "month")
		if err != nil {
			panic(err)
		}
		err = monthElem.SendKeys("July")
		if err != nil {
			log.Println("Month wasn't entered.")
		} else {
			log.Println("Month entered.")
		}

		dayElem, err := wd.FindElement(selenium.ByID, "day")
		if err != nil {
			panic(err)
		}
		err = dayElem.SendKeys("1")
		if err != nil {
			log.Println("Day wasn't entered.")
		} else {
			log.Println("Day entered.")
		}

		yearElem, err := wd.FindElement(selenium.ByID, "year")
		if err != nil {
			panic(err)
		}
		err = yearElem.SendKeys("1998")
		if err != nil {
			log.Println("Year wasn't entered.")
		} else {
			log.Println("Year entered.")
		}

		// Random Gender Selection
		minGender := 1
		maxGender := 3
		randomValGender := rand.Intn(maxGender - minGender) + minGender

		switch randomValGender {
		case 1:
			male, err := wd.FindElement(selenium.ByCSSSelector, "span[class='Indicator-sc-16vj7o8-0 dDbCKU']")
			if err != nil {
				panic(err)
			}
			if err := male.Click(); err != nil {
				_,_ = red.Println("\n\tMale gender button not clicked.\n")
			} else {
				log.Println("Male gender button clicked.")
			}
			break
		case 2:
			female, err := wd.FindElement(selenium.ByCSSSelector, "span[class='Indicator-sc-16vj7o8-0 dDbCKU']")
			if err != nil {
				panic(err)
			}
			if err := female.Click(); err != nil {
				_,_ = red.Println("\n\tFemale gender button not clicked.\n")
			} else {
				log.Println("Female gender button clicked.")
			}
			break
		case 3:
			neutral, err := wd.FindElement(selenium.ByCSSSelector, "span[class='Indicator-sc-16vj7o8-0 dDbCKU']")
			if err != nil {
				panic(err)
			}
			if err := neutral.Click(); err != nil {
				_,_ = red.Println("\n\tNon-binary gender button not clicked.\n")
			} else {
				log.Println("Non-binary gender button clicked.")
			}
			break
		default:
			neutral, err := wd.FindElement(selenium.ByID, "register-neutral")
			if err != nil {
				panic(err)
			}
			if err := neutral.Click(); err != nil {
				_,_ = red.Println("\n\tNon-binary gender button not clicked.\n")
			} else {
				log.Println("Non-binary gender button clicked.")
			}
			break
		}

		v2Solver(wd)
	}
}

func v3Solver(wd selenium.WebDriver, v3Enabled bool) {
	if v3Enabled {
		c := Captcha.New(twoCaptchaAPIKey)

		solved, err := c.SolveRecaptchaV3(SpotifyRegistrationURL, v3ReCaptchaKey, "t", "0.3")
		if err != nil {
			wd.Quit()
			SpotifyInit()

			_ = err
		} else {
			log.Println("[✓](v3) Solved via 2captcha.com, send back to site..") // String

			// Send Solved Key
			_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response-100000').innerHTML='" + solved + "';"), nil)
			if err != nil {
				panic(fmt.Sprintf("[✕](v3) Reponse Key Submission Error: %s", err)) // ReCaptcha Key wasn't submitted back to website.
			} else {
				log.Println("[✓](v3) ReCaptcha Response Key submitted back to target site.")
			}

			_ = solved
		}
	}
}

func v2Solver(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(v2Solver) Error: ", r)

			wd.Quit() // Close previous session
			time.Sleep(10 * time.Second) // Timeout
			initSpotifyAccountCreation(wd) // Restart
		}
	}()

	cV2 := Captcha.New(twoCaptchaAPIKey)

	solved, err := cV2.SolveRecaptchaV2(SpotifyRegistrationURL, v2ReCaptchaKey)
	if err != nil {
		panic(err)
	} else {
		log.Println("[✓](v2) Solved via 2captcha.com.") // String

		// Send Solved Key back to site recaptcha
		_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response').innerHTML='" + solved + "';"), nil)
		if err != nil {
			panic(fmt.Sprintf("[✕](v2) Reponse Key Submission Error: %s", err)) // ReCaptcha Key wasn't submitted back to website.
		} else {
			log.Println("[✓](v2) ReCaptcha Response Key submitted back to Spotify.")
		}
		
		// Callback Execution
		_, err = wd.ExecuteScript(fmt.Sprintf("___grecaptcha_cfg.clients[0]%scallback('%s')", captchaCallback, solved), nil) // Chrome Inspect > Console > Type '___grecaptcha_cfg.clients[0].' until callback discovered
		if err != nil {
			panic(fmt.Sprintf("[✕](v2) Callback not executed: %s", err)) // ReCaptcha Callback function wasn't executed
		} else {
			log.Println("[✓](v2) ReCaptcha Callback executed.")

			// Submit Form via Register button
			_, err = wd.ExecuteScript(fmt.Sprintf("document.querySelector('.Button-oyfj48-0.eEPJyH.SignupButton___StyledButtonPrimary-cjcq5h-1.deUbNh').click();"), nil)
			if err != nil {
				panic(fmt.Sprintf("[✕](v2) Submit button not clicked: %s", err)) // ReCaptcha Form wasn't submitted.
			} else {
				log.Println("[✓](v2) Submit button clicked.")

				time.Sleep(5 * time.Second) // Wait

				// Error Message Check
				_, err := wd.FindElement(selenium.ByCSSSelector, recaptchaError)
				if err != nil {
					log.Println("[✓](v2) ReCaptcha successfully solved!")
				} else {
					_,_ = red.Println(fmt.Sprintf("\n\t[✕](v2) 'Confirm you're not a robot' msg dislayed!"))

					restartSpotify(wd) // Restart
				}

				registrationCheckSpotify(wd)
			}
		}
	}
}

func registrationCheckSpotify(wd selenium.WebDriver) {
	time.Sleep(5 * time.Second) // Wait

	_, err := wd.FindElement(selenium.ByCSSSelector, robotErrorLabel)
	if err != nil {
		// Double check credentials
		_, err := wd.FindElement(selenium.ByXPATH, yourLibraryNavBtn)
		if err != nil {
			_,_ = cyan.Println(fmt.Sprintf("\n\tUser successfully registered! [NewCreatedAccount #%d]\n", accountCreationCount))

			accountCreationCount++
			Database.AddSpotifyAccount(UserGenEmailSpotifyGlbl, UserGenPasswordSpotifyGlbl) // ADD credentials to database
			_,_ = magenta.Println("\tNew credentials added to remote DB!")

			accountsInDB = accountsInDB + 1 // Sync DB call

			// Create required number of accounts before streaming again
			if accountCreationCount >= accountsRequired {
				// Check if enough accounts have been created
				if accountsInDB >= accountsRequired {
					restartSpotify(wd)
				} else {
					wd.Quit()
					initSpotifyAccountCreation(wd) // Continue to create accounts
				}
			} else {
				wd.Quit()
				initSpotifyAccountCreation(wd) // Continue to create accounts
			}
		}
	} else {
		accountCreationErrorCount++

		_,_ = red.Println("\n\tUser not added to remote DB, refreshing ...")

		if accountCreationErrorCount <= 3 {
			accountCreationErrorCount = 0

			wd.Refresh()
			registrationCheckSpotify(wd)
		} else {
			accountCreationErrorCount = 0

			wd.Quit()
			initSpotifyAccountCreation(wd)
		}
	}
}

func misplacedTrackCheck(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(shuffleSpotify) panic occured: ", r)

			wd.Refresh()
			misplacedTrackCheck(wd)
		}
	}()

	trackTitleDiv, err := wd.FindElement(selenium.ByCSSSelector, trackTitleDiv)
	if err != nil {
		panic(err)
	}

	trackTitle, _ := trackTitleDiv.Text()

	if trackTitle == "Just Saying(feat.Baby G)" { // Forbid this song from obtaining streams
		skipBtn, err := wd.FindElement(selenium.ByCSSSelector, skipBtn)
		if err != nil {
			panic(err)
		}

		if err := skipBtn.Click(); err != nil {
			misplacedTrackCheckFail++
			_, _ = red.Printf("\tSkip button click failed.\n")

			if misplacedTrackCheckFail % 3 == 0 {
				misplacedTrackCheckFail = 0 // Reset

				restartSpotify(wd)
			}
		} else {
			log.Printf("#(%d) Misplaced track '%s' skipped.", shuffleIterationSpotify, trackTitle)
		}
	}
}

func estimatedRoyaltyCalc() {
	newEst := estimatedFigCount

	_, _ = green.Println(fmt.Sprintf("\tEstimation of Royalties: ~$%.2f", newEst))
}

func pricePerStreamGen(x float64, y float64) string {
	// Used to get an estimated average price per stream

	f := randPriceFloats(x, y, 1)
	s := fmt.Sprintf("%f", f) // Convert float64[] to string
	a := strings.Replace(s, "[", "", -1) // Trim [ from string
	b := strings.Replace(a, "]", "", -8) // Trim ] from string
	c, _ := strconv.ParseFloat(b, 64) // Convert string to float
	d := fmt.Sprintf("%.5f", c)

	estimatedFigCount = estimatedFigCount + c

	return d
}

func randPriceFloats(min, max float64, n int) []float64 {
	rand.Seed(time.Now().UTC().UnixNano())

	res := make([]float64, n)

	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}

	return res
}
