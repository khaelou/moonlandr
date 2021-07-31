package Worker

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	Database "moonlandr/db"
	Fetch "moonlandr/fetch"
	Captcha "moonlandr/twocaptchaV3"

	"github.com/fatih/color"
	godotenv "github.com/joho/godotenv"
	conditions "github.com/serge1peshcoff/selenium-go-conditions"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var (
	green   = color.New(color.FgHiGreen, color.Bold)
	cyan    = color.New(color.FgHiCyan, color.Bold)
	yellow  = color.New(color.FgYellow, color.Bold)
	red     = color.New(color.FgRed, color.Bold)
	magenta = color.New(color.FgMagenta, color.Bold)

	accountsRequired int
	streamIterations int
	playbackWaitTime int
	host             string
	captchaCallback  string
	enableVNC        bool
	enableV3         bool

	UserIdSpotify              int
	UserEmailSpotify           string
	UserPasswordSpotify        string
	UserGenEmailSpotifyGlbl    string
	UserGenPasswordSpotifyGlbl string

	ArtistProfileSpotify string
	ArtistNameSpotify    string
	mlPlaybackTime       string

	holdItTime                = time.Duration(4)
	accountsInDB              = 0
	accountSigninErrorCount   = 0
	accountCreationCount      = accountsInDB + 1
	accountCreationErrorCount = 0
	misplacedTrackCheckFail   = 0
	shuffleIterationSpotify   = 1
	shuffleInvalidations      = 0
	estimatedFigCount         = 0.00000

	// Simply modify attributes here if Spotify where to make HTML changes
	incorrectLoginAlert = "//p[@class='alert alert-warning']"
	artistNameLabel     = "//h1[@class='a12b67e576d73f97c44f1f37026223c4-scss']"
	signupBtn           = "//button[@class='Button-sc-8cs45s-0 jgLyVk']"
	followBtn           = "//button[@class='ff6a86a966a265b5a51cf8e30c6c52f4-scss']"
	unfollowBtn         = "//button[@class='ff6a86a966a265b5a51cf8e30c6c52f4-scss _888a8dffe06d27b161f0258c2769069e-scss']"
	startDiscographyBtn = "//button[@class='_8e7d398e09c25b24232d92aac8a15a81-scss e8b2fe03d4e4726484b879ed8ff6f096-scss']"
	playBtn             = "//button[@data-testid='play-button' and not(@disabled)]"
	loadingPlayBtn      = "//button[@class='_82ba3fb528bb730b297a91f46acd37a3-scss']"
	skipBtn             = "//button[@class='bf01b0d913b6bfffea0d4ffd7393c4af-scss']"
	disabledSkipBtn     = "//button[@class='bf01b0d913b6bfffea0d4ffd7393c4af-scss']"
	favTrackBtn         = "//button[@class='_07bed3a434fa59aa1852a431bf2e19cb-scss control-button control-button-heart']"
	unFavTrackBtn       = "//button[@class='_07bed3a434fa59aa1852a431bf2e19cb-scss control-button control-button-heart a65d8d62fe56eed3e660b937a9be8a93-scss']"
	playbackTimeDiv     = "//div[@class='playback-bar__progress-time _5f899d811cf206c5925f6450626fb0aa-scss']"
	trackTitleDiv       = "//div[@class='_86f3bde5c3f38a2a37d03381c41acaf4-scss ellipsis-one-line f3fc214b257ae2f1d43d4c594a94497f-scss']"
	artistNameDiv       = "//div[@class='f9ac49a03051d20affdc7135cfdbad3e-scss ellipsis-one-line _5f899d811cf206c5925f6450626fb0aa-scss']"
	accMenuBtn          = "//button[@class='_3e75c7f07bdce28b37b45a5cd74ff371-scss']"
	logoutBtn           = "//button[@class='d2a8e42f26357f2d21c027f30d93fb64-scss']"
	isSignUpDone        = "//div[@class='ButtonInner-peijbp-0 drgjVo FacebookButton__StyledFacebookButton-sc-4xbei5-1 IsOOA']"
	//pauseBtn        = "//button[@class='_82ba3fb528bb730b297a91f46acd37a3-scss' or @title='Pause' and not(@disabled)]"
	//premiumModal    = "//div[@class='GenericModal GenericModal--animated _9503df1e6a7a900ae17aeba014203575-scss GenericModal--afterOpen']"
	//premiumModalBtn = "//button @class='Button-sc-1dqy6lx-0 cLnKJb _1202545091238e5aa5b47b15ab6786fe-scss e810fe421a0b204c0de3771cf655e135-scss']"

	twoCaptchaAPIKey string
	v2ReCaptchaKey   string // Pulled from 'SpotifyRegistrationURL'
	v3ReCaptchaKey   string // Pulled from 'SpotifyRegistrationURL'
)

const (
	port                   = 4444                                                                               // 4444 (Selenoid Port)
	SpotifyRegistrationURL = "https://www.spotify.com/us/signup/?forward_url=https%3A%2F%2Fopen.spotify.com%2F" // For account creation
)

func initArtistSpotify() {
	Database.SpotifyArtistConnectDB() // Open remote DB connection for artists
	ArtistProfileSpotify = Database.ReturnSpotifyArtistPortal()
	ArtistNameSpotify = Database.ReturnSpotifyArtistTitle()

	fmt.Printf("\n\tSTREAM CLIENT: " + ArtistNameSpotify + "\n")

	Database.SpotifyAccountsConnectDB() // Open remote DB connection for accounts
	Fetch.Creds()                       // Retrieve selected Spotify credentials for the session
	UserIdSpotify = Fetch.SpotifyId
	UserEmailSpotify = Fetch.SpotifyEmail
	UserPasswordSpotify = Fetch.SpotifyPassword
}

func restartSpotify(wd selenium.WebDriver) {
	_, _ = yellow.Println("Restarting worker ...")

	wd.Quit() // Quit previous session
	SpotifyInit()
}

func hostBalancerSpotify() {
	host1 := Database.ReturnSelenoidHost()  // Location 1
	host2 := Database.ReturnSelenoidHost2() // Location 2

	selectedHost := []string{
		host1,
		host2,
	}

	rand.Seed(time.Now().UnixNano())

	host = selectedHost[rand.Intn(len(selectedHost))]

	switch host {
	case host1:
		fmt.Println("\n\t— Selenoid Host 1:", fmt.Sprintf("%s:%d/wd/hub", host, port))
	case host2:
		fmt.Println("\n\t— Selenoid Host 2:", fmt.Sprintf("%s:%d/wd/hub", host, port))
	}
}

func SpotifyInit() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	twoCaptchaAPIKey = os.Getenv("2CAPTCHA_API_KEY")
	v2ReCaptchaKey = os.Getenv("V2_RECAPTCHA_KEY")
	v3ReCaptchaKey = os.Getenv("V3_RECAPTCHA_KEY")

	Database.PullSpotifySettingsConnectDB()            // Open remote DB connection for config settings
	accountsInDB = Database.ReturnNumSpotifyAccounts() // Placed here to refresh value if more accounts are added
	accountsRequired = Database.ReturnAccountsRequired()
	streamIterations = Database.ReturnStreamIterations()
	playbackWaitTime = Database.ReturnPlayBackWaitTime()
	captchaCallback = Database.ReturnReCaptchaCallBack()
	enableVNC = Database.ReturnEnableVNC()
	enableV3 = Database.ReturnEnableV3()

	hostBalancerSpotify() // Used to select a Selenoid Host, better manage CPU

	// Check if enough accounts are in remote DB
	if accountsInDB >= accountsRequired {
		fmt.Println("\n\t[WebPlayer] Browser >>", "Chrome 91.0")
		fmt.Println("\n\t* ]", accountsInDB, "queued accounts in remote DB.")

		initArtistSpotify()                                                       // Pull credentials from remote DB
		loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify) // Authenticate Spotify account
	} else {
		fmt.Println("\n\t[Signup] Browser >>", "Chrome 91.0")

		// Connect to the WebDriver instance (Selenoid)
		caps := selenium.Capabilities{"browserName": "chrome", "version": "91.0", "enableVNC": enableVNC, "useAutomationExtension": false}
		chromeCaps := chrome.Capabilities{
			Args: []string{
				"start-maximized",
				"--disable-blink-features",
				"--disable-blink-features=AutomationControlled",
				"--disable-crash-reporter",
				"--ignore-certifcate-errors",
				"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.53 Safari/537.36",
			},
			ExcludeSwitches: []string{
				"enable-automation",
			},
		}
		caps.AddChrome(chromeCaps)

		wd, err := selenium.NewRemote(caps, fmt.Sprintf("%s:%d/wd/hub", host, port))
		if err != nil {
			panic(err)
		}
		defer wd.Quit()
		wd.MaximizeWindow("")

		initSpotifyAccountCreation(wd)
	}
}

func loginProcessSpotify(id int, email, password string) {
	// Connect to the WebDriver instance (Selenoid)
	caps := selenium.Capabilities{"browserName": "chrome", "version": "91.0", "enableVNC": enableVNC, "useAutomationExtension": false}
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"start-maximized",
			"--disable-blink-features",
			"--disable-blink-features=AutomationControlled",
			"--disable-crash-reporter",
			"--ignore-certifcate-errors",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.53 Safari/537.36",
		},
		ExcludeSwitches: []string{
			"enable-automation",
		},
	}
	caps.AddChrome(chromeCaps)

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("%s:%d/wd/hub", host, port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()
	wd.MaximizeWindow("") // Maximize Window regardless of Browser

	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(loginProcessSpotify) panic occured:", r)

			wd.Quit()
			time.Sleep(holdItTime * time.Second) // Wait

			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}()

	fmt.Println("\tSessionID:", wd.SessionID()) // Selenoid UI

	fmt.Println("\n\tACC ID:", id)
	fmt.Printf("\n\tACCOUNT EMAIL: " + email)
	fmt.Printf("\n\tACCOUNT PASSWORD: " + password + "\n")
	fmt.Println()

	// Additional check verifying if the Remote DB has any accounts stored, if not create some
	if id == 0 && email == "0" && password == "0" {
		initSpotifyAccountCreation(wd)
	}

	// Navigate to the (Login Page -> Artist Profile)
	if err := wd.Get(ArtistProfileSpotify); err != nil {
		panic(err)
	} else {
		_, _ = yellow.Println("\tPage reached.")

		if title, err := wd.Title(); err != nil {
			_, _ = red.Printf("\n\tFailed to get page title: %s\n", err)
		} else {
			fmt.Printf("\n\t%s\n\n", title)
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
		_, _ = red.Println("\n\tLogin button not clicked.\n")
	} else {
		log.Println("Login button clicked.")

		credentialCheckSpotify(wd)
	}
}

func credentialCheckSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(credentialCheckSpotify) panic occured:", r)

			wd.Quit()
			time.Sleep(holdItTime * time.Second) // Wait

			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}()

	time.Sleep(holdItTime * time.Second) // Wait

	// Check for login error
	_, err := wd.FindElement(selenium.ByXPATH, incorrectLoginAlert)
	if err != nil {
		time.Sleep(holdItTime * time.Second) // Wait

		// Double check credentials
		_, err := wd.FindElement(selenium.ByXPATH, artistNameLabel)
		if err != nil {
			_, _ = red.Println("\n\tUser not authenticated, restart...\n", err)
			panic(err)
		} else {
			_, _ = cyan.Println("\n\tUser successfully authenticated!\n")

			time.Sleep(holdItTime * time.Second) // Wait

			// Loading Play button check
			_, err := wd.FindElement(selenium.ByXPATH, loadingPlayBtn)
			if err != nil {
				log.Println("Loading play button:", err)

				// Check if play button present
				playerPlayBtn, err := wd.FindElement(selenium.ByXPATH, playBtn)
				if err != nil {
					log.Println("Play button not located:", err)
					panic(err)
				}
				if err := playerPlayBtn.Click(); err != nil {
					log.Println("Play button not clicked:", err)
					panic(err)
				} else {
					_, _ = yellow.Println("\tFinally playing discography!\n")
				}
			} else {
				// Unfollow button check
				_, err := wd.FindElement(selenium.ByXPATH, unfollowBtn)
				if err != nil {
					followArtistSpotify(wd)
					startDiscographySpotify(wd)
				} else {
					startDiscographySpotify(wd)
				}
			}
		}
	} else {
		_, _ = red.Println("\tInvalid Authentication:", UserEmailSpotify)

		accountSigninErrorCount++

		if accountSigninErrorCount%4 == 0 {
			accountSigninErrorCount = 0

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
		} else {
			wd.Quit()
			time.Sleep(holdItTime * time.Second) // Wait

			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}
}

func followArtistSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("(followArtistSpotify) panic occured:", r)

			logoutSpotify(wd)
		}
	}()

	time.Sleep(holdItTime * time.Second) // Wait

	followBtn, err := wd.FindElement(selenium.ByXPATH, followBtn)
	if err != nil {
		panic(err)
	}
	if err := followBtn.Click(); err != nil {

		panic(err)
	} else {
		_, _ = green.Println("\tNow following artist!\n")
	}
}

func favoriteTrackSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(favoriteTrackSpotify) panic occured: Like button not visible.")
			_ = r
		}
	}()

	trackTitleDiv, err := wd.FindElement(selenium.ByXPATH, trackTitleDiv)
	if err != nil {
		panic(err)
	} else {
		trackTitle, _ := trackTitleDiv.Text()

		artistNameDiv, err := wd.FindElement(selenium.ByXPATH, artistNameDiv)
		if err != nil {
			log.Println("artistNameDiv:", err)
		} else {
			artistName, _ := artistNameDiv.Text()

			// Check if song already favorited
			_, err := wd.FindElement(selenium.ByXPATH, unFavTrackBtn)
			if err != nil {
				favBtn, err := wd.FindElement(selenium.ByXPATH, favTrackBtn)
				if err != nil {
					panic(err)
				}
				if err := favBtn.Click(); err != nil {
					_, _ = red.Println("\tLike failed:", err)
				} else {
					_, _ = cyan.Println(fmt.Sprintf("\tTrack '%s' by '%s' has been liked!", trackTitle, artistName))
				}
			} else {
				_, _ = magenta.Println(fmt.Sprintf("\tTrack '%s' by '%s' pre-liked!", trackTitle, artistName))
			}
		}
	}
}

func startDiscographySpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); nil != r {
			_, _ = red.Println("(startDiscographySpotify) panic occured:", r)

			wd.Quit()
			time.Sleep(holdItTime * time.Second) // Wait

			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}()

	time.Sleep(holdItTime * time.Second) // Wait

	// Pause button check
	_, err := wd.ExecuteScript("document.getElementsByClassName('_8e7d398e09c25b24232d92aac8a15a81-scss e8b2fe03d4e4726484b879ed8ff6f096-scss')[0]", nil)
	if err != nil {
		_, _ = red.Println("Pause button visible:", err)
	} else {
		time.Sleep(holdItTime * time.Second) // Wait

		_, err = wd.ExecuteScript("document.getElementById('onetrust-consent-sdk').remove();", nil)
		if err != nil {
			_, _ = red.Println("\n\tObscurity `onetrust-consent-sdk` not removed", err)
		} else {
			time.Sleep(holdItTime * time.Second) // Delay Wait

			_, err = wd.ExecuteScript("document.getElementsByClassName('contentSpacing _4c3b6e4e88112fc8ef88512cbe7521ed-scss da51a6e223c7200d373a2fd0614d7c33-scss')[0].remove();", nil)
			if err != nil {
				_, _ = red.Println("\n\tObscurity #2 not removed:", err)
				panic(err)
			} else {
				if err := wd.Wait(conditions.ElementIsLocatedAndVisible(selenium.ByXPATH, startDiscographyBtn)); err != nil {
					log.Println("Discography start button not located:", err)
				} else {
					startBtn, err := wd.FindElement(selenium.ByXPATH, startDiscographyBtn)
					if err != nil {
						log.Println("startBtn:", err)
					}
					if err := startBtn.Click(); err != nil {
						_, _ = red.Println("\n\tDiscography start failed:", err)
					} else {
						log.Println("Now Streaming w/", time.Duration(playbackWaitTime)*time.Second, "playback-duration!")

						SpotifyRepeatShuffleEvery(time.Duration(playbackWaitTime)*time.Second, wd)
					}
				}
			}
		}
	}
}

func SpotifyRepeatShuffleEvery(d time.Duration, wd selenium.WebDriver) {
	defer func() {
		if r := recover(); nil != r {
			_, _ = red.Println("SpotifyRepeatShuffleEvery('Playback stand-still') panic occured:", r)

			wd.Quit()
			time.Sleep(holdItTime * time.Second) // Wait

			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}()

	if err := wd.Wait(conditions.ElementIsLocatedAndVisible(selenium.ByXPATH, artistNameDiv)); err != nil {
		_, _ = red.Println("Playback stand-still:", err)
		panic(err)
	} else {
		artistNameDiv, err := wd.FindElement(selenium.ByXPATH, artistNameDiv)
		if err != nil {
			log.Println("artistNameDiv:", err)
		} else {
			artistName, _ := artistNameDiv.Text()

			log.Println(fmt.Sprintf("(#%d) "+ArtistNameSpotify+"'s Discography Started!", shuffleIterationSpotify))

			trackTitleDiv, err := wd.FindElement(selenium.ByXPATH, trackTitleDiv)
			if err != nil {
				panic(err)
			} else {
				trackTitle, _ := trackTitleDiv.Text()

				if artistName != ArtistNameSpotify {
					_, _ = yellow.Printf("\n\t[LayOver] '" + trackTitle + "' by " + artistName + " has resumed ...\n\n")
				} else {
					_, _ = yellow.Printf("\n\t[LayOver] '" + trackTitle + "' by " + ArtistNameSpotify + " has resumed ...\n\n")
				}
			}
		}

		for _ = range time.Tick(d) {
			shuffleSpotify(wd)
		}
	}
}

func shuffleSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(shuffleSpotify) panic occured:", r)

			wd.Quit()
			time.Sleep(holdItTime * time.Second) // Wait

			loginProcessSpotify(UserIdSpotify, UserEmailSpotify, UserPasswordSpotify)
		}
	}()

	time.Sleep(holdItTime * time.Second)

	// Advertisement check
	_, err := wd.FindElement(selenium.ByXPATH, disabledSkipBtn)
	if err != nil {
		_, _ = yellow.Println("Advertisement:", err)

		panic(err)
	} else {
		playbackTimeDiv, err := wd.FindElement(selenium.ByXPATH, playbackTimeDiv)
		if err != nil {
			_, _ = red.Println("\tPlayback Duration Error:", err)
			logoutSpotify(wd)
		}

		playbackTime, _ := playbackTimeDiv.Text()
		minuteFromTime := playbackTime[0:1] // removes ":00" from playback time
		minute, _ := strconv.ParseInt(minuteFromTime, 10, 64)
		secondsFromTime := playbackTime[2:] // removes "0:" from playback time | 2:len(playbackTime)
		seconds, _ := strconv.ParseInt(secondsFromTime, 10, 64)

		if seconds >= 10 {
			mlPlaybackTime = strconv.Itoa(int(minute)) + ":" + strconv.Itoa(int(seconds))
		} else {
			mlPlaybackTime = strconv.Itoa(int(minute)) + ":0" + strconv.Itoa(int(seconds))
		}

		if playbackTime == mlPlaybackTime {
			mlPlaybackTime = playbackTime
		} else {
			_, _ = red.Println("\n\tPlayback Duration Read Error.\n")
		}

		trackTitleDiv, err := wd.FindElement(selenium.ByXPATH, trackTitleDiv)
		if err != nil {
			logoutSpotify(wd)
		}
		trackTitle, _ := trackTitleDiv.Text()

		// Shuffle button
		shuffleBtn, err := wd.FindElement(selenium.ByXPATH, skipBtn)
		if err != nil {
			panic(err)
		}
		if err := shuffleBtn.Click(); err != nil {
			_, _ = red.Println("\n\tShuffle button click failed:", err)
		} else {
			if seconds > 0 {
				// Check if playback surpassed 30 seconds
				if seconds >= 30 && minute <= 1 || seconds < 30 && minute > 0 {
					artistNameDiv, err := wd.FindElement(selenium.ByXPATH, artistNameDiv)
					if err != nil {
						log.Println("artistNameDiv:", err)
					} else {
						artistName, _ := artistNameDiv.Text()

						if artistName != ArtistNameSpotify {
							log.Printf("[✓] Track '%s' by '%s' playback stopped at %s. [~$%s]", trackTitle, artistName, mlPlaybackTime, pricePerStreamGenSpotify(0.00331, 0.00437))
						} else {
							log.Printf("[✓] Track '%s' by '%s' playback stopped at %s. [~$%s]", trackTitle, ArtistNameSpotify, mlPlaybackTime, pricePerStreamGenSpotify(0.00331, 0.00437))
						}

						estimatedRoyaltyCalcSpotify()

						// Shuffle/Skip button
						skipBtn, err := wd.FindElement(selenium.ByXPATH, skipBtn)
						if err != nil {
							panic(err)
						}
						if err := skipBtn.Click(); err != nil {
							shuffleInvalidations++
							_, _ = red.Printf("\tShuffle to next song failed. #(%d)\n", shuffleInvalidations)

							if shuffleInvalidations%3 == 0 {
								shuffleInvalidations = 0
								panic(err)
							}
						} else {
							shuffleIterationSpotify++

							log.Printf("#(%d) Shuffle to next song.", shuffleIterationSpotify)

							misplacedTrackCheckSpotify(wd) // Check if there is a misplaced track before continuing
							favoriteTrackSpotify(wd)
						}
					}
				}
			}
		}
	}

	// Iteration check
	if shuffleIterationSpotify%streamIterations == 0 {
		_, _ = red.Println("\n\tStream limit reached, rotating..")

		logoutSpotify(wd)
	}
}

func logoutSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(logoutSpotify) panic occured: ", r)

			restartSpotify(wd)
		}
	}()

	menuBtn, err := wd.FindElement(selenium.ByXPATH, accMenuBtn)
	if err != nil {
		panic(err)
	}
	if err := menuBtn.Click(); err != nil {
		_, _ = red.Println("Account Menu click failed!")
	} else {
		logoutBtn, err := wd.FindElement(selenium.ByXPATH, logoutBtn)
		if err != nil {
			panic(err)
		}

		if err := logoutBtn.Click(); err != nil {
			_, _ = red.Println("\n\tLogout button click failed.\n")
		} else {
			_, _ = yellow.Println("\n\tSuccessfully logged out.")

			restartSpotify(wd)
		}
	}
}

func initSpotifyAccountCreation(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(initSpotifyAccountCreation) panic occured:", r)

			restartSpotify(wd)
		}
	}()

	fmt.Println("\tSessionID:", wd.SessionID()) // Selenoid UI

	minWait := 5
	maxWait := 15
	holdItTime = time.Duration(rand.Intn(maxWait-minWait) + minWait)
	fmt.Println("\n\ttime.Sleep() Duration:", holdItTime)

	// Check if enough accounts are in remote DB
	if accountsInDB >= accountsRequired {
		restartSpotify(wd)
	} else {
		fmt.Println("\n\t* ]", accountsInDB, "queued accounts in remote DB. [ACCOUNT CREATOR]")

		rand.Seed(time.Now().UTC().UnixNano())
		min := 1
		max := 99999999
		randomVal := rand.Intn(max-min) + min

		UserGenEmailSpotify := fmt.Sprintf("%s%d%s", "landrmoon", randomVal, "@protonmail.com")
		UserGenPasswordSpotify := "SpaceCadet1231!"
		UserGenNameSpotify := fmt.Sprintf("%s%d", "betalandr", randomVal)

		UserGenEmailSpotifyGlbl = UserGenEmailSpotify
		UserGenPasswordSpotifyGlbl = UserGenPasswordSpotify

		// Navigate to the (Registration Page)
		if err := wd.Get(SpotifyRegistrationURL); err != nil {
			panic(err)
		} else {
			_, _ = yellow.Println("\n\tPage reached.")

			if title, err := wd.Title(); err != nil {
				_, _ = red.Printf("\n\tFailed to get page title: %s\n", err)
			} else {
				fmt.Printf("\n\t%s\n\n", title)
			}
		}

		// Bypass 'navigator.webdriver'
		_, err := wd.ExecuteScript("Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); document.getElementsByClassName('Type__TypeElement-sc-9snywk-0 bRyGwI')[0].innerHTML=navigator.webdriver;", nil)
		if err != nil {
			panic(err)
		} else {
			retrieveNavigatorWebDriver := "return document.getElementsByClassName('Type__TypeElement-sc-9snywk-0 bRyGwI')[0].innerHTML"
			jsRetrieveNavigatorWebDriver, err := wd.ExecuteScript(retrieveNavigatorWebDriver, nil)
			if err != nil {
				panic(err)
			} else {
				_, _ = cyan.Println("OK (navigator.webdriver):", jsRetrieveNavigatorWebDriver)

				v3SolverSpotify(wd, enableV3)

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

				time.Sleep(holdItTime * time.Second) // Wait

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

				time.Sleep(holdItTime * time.Second) // Wait

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

				time.Sleep(holdItTime * time.Second) // Wait

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

				time.Sleep(holdItTime * time.Second) // Wait

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

				time.Sleep(holdItTime * time.Second) // Wait

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

				time.Sleep(holdItTime * time.Second) // Wait

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

				time.Sleep(holdItTime * time.Second) // Wait

				// Scroll Down
				scrollDownSpotify(wd)

				// Random Gender Selection
				minGender := 1
				maxGender := 4
				randomValGender := rand.Intn(maxGender-minGender) + minGender

				switch randomValGender {
				case 1:
					_, err = wd.ExecuteScript("document.getElementById('gender_option_male').click();", nil)
					if err != nil {
						_, _ = red.Println("\n\tMale gender button not clicked.\n")
						panic(err)
					} else {
						log.Println("Male gender button clicked.")
					}
				case 2:
					_, err = wd.ExecuteScript("document.getElementById('gender_option_female').click();", nil)
					if err != nil {
						_, _ = red.Println("\n\tFemale gender button not clicked.\n")
						panic(err)
					} else {
						log.Println("Female gender button clicked.")
					}
				case 3:
					_, err = wd.ExecuteScript("document.getElementById('gender_option_nonbinary').click();", nil)
					if err != nil {
						_, _ = red.Println("\n\tNon-binary gender button not clicked.\n")
						panic(err)
					} else {
						log.Println("Non-binary gender button clicked.")
					}
				default:
					_, err = wd.ExecuteScript("document.getElementById('gender_option_nonbinary').click();", nil)
					if err != nil {
						_, _ = red.Println("\n\tNon-binary gender button not clicked.\n")
						panic(err)
					} else {
						log.Println("Non-binary gender button clicked.")
					}
				}

				_, err = wd.ExecuteScript("document.getElementById('onetrust-consent-sdk').remove();", nil)
				if err != nil {
					_, _ = red.Println("\n\tObsurity `onetrust-consent-sdk` not removed", err)
				} else {
					time.Sleep(holdItTime * time.Second) // Wait
					v2SolverSpotify(wd)
				}
			}
		}
	}
}

func v3SolverSpotify(wd selenium.WebDriver, v3Enabled bool) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(v3SolverSpotify) panic occured: ", r)

			wd.Refresh()
			initSpotifyAccountCreation(wd) // Restart
		}
	}()

	if v3Enabled {
		c := Captcha.New(twoCaptchaAPIKey)

		solved, err := c.SolveRecaptchaV3(SpotifyRegistrationURL, v3ReCaptchaKey, "t", "0.3")
		if err != nil {
			panic(err)
		} else {
			log.Println("[✓](v3) Solved via 2captcha.com, send back to site..") // String

			// Send Solved Key
			_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response-100000').innerHTML='"+solved+"';"), nil)
			if err != nil {
				panic(fmt.Sprintf("[✕](v3) Reponse Key Submission Error: %s", err)) // ReCaptcha Key wasn't submitted back to website.
			} else {
				log.Println("[✓](v3) ReCaptcha Response Key submitted back to target site.")
			}

			_ = solved
		}
	}
}

func v2SolverSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(v2SolverSpotify) panic occured: ", r)

			wd.Refresh()
			initSpotifyAccountCreation(wd) // Restart
		}
	}()

	log.Println("v2 Captcha Bypass ...")

	cV2 := Captcha.New(twoCaptchaAPIKey)

	solved, err := cV2.SolveRecaptchaV2(SpotifyRegistrationURL, v2ReCaptchaKey)
	if err != nil {
		panic(err)
	} else {
		log.Println("[✓](v2) Solved via 2captcha.com:", solved) // String

		// Show g-recaptcha-response TextArea for debugging
		_, err = wd.ExecuteScript("document.querySelector('#g-recaptcha-response').style.removeProperty('display');", nil)
		if err != nil {
			panic(fmt.Sprintf("[✕] g-recaptcha-response TextArea not invisible : %s", err)) // ReCaptcha Key wasn't submitted back to website.
		} else {
			log.Println("[✓] g-recaptcha-response now visible.")
		}

		// Send Solved Key back to site recaptcha
		_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response').innerHTML='"+solved+"';"), nil)
		if err != nil {
			panic(fmt.Sprintf("[✕](v2) Reponse Key Submission Error: %s", err)) // ReCaptcha Key wasn't submitted back to website.
		} else {
			log.Println("[✓](v2) ReCaptcha Response Key submitted back to Spotify.")
		}

		// Show g-recaptcha-response-100000 TextArea for debugging
		_, err = wd.ExecuteScript("document.querySelector('#g-recaptcha-response-100000').style.removeProperty('display');", nil)
		if err != nil {
			panic(fmt.Sprintf("[✕] g-recaptcha-response-100000 TextArea not invisible : %s", err)) // ReCaptcha Key wasn't submitted back to website.
		} else {
			log.Println("[✓] g-recaptcha-response-100000 now visible.")
		}

		// Send Solved Key back to site recaptcha response-100000 reference
		_, err = wd.ExecuteScript(fmt.Sprintf("document.getElementById('g-recaptcha-response-100000').innerHTML='"+solved+"';"), nil)
		if err != nil {
			panic(fmt.Sprintf("[✕✕](v2-100000) Reponse Key Submission Error: %s", err)) // ReCaptcha Key wasn't submitted back to website.
		} else {
			log.Println("[✓✓](v2-100000) ReCaptcha Response Key submitted back to Spotify.")
		}

		// Find v2 '[0].callback' Captcha Callback, add value to a <p> tag via class name for retrieval
		discoverV2Clients := "function findRecaptchaClients(){if(typeof(___grecaptcha_cfg)!=='undefined'){return Object.entries(___grecaptcha_cfg.clients).map(([cid,client])=>{const data={id:cid,version:cid>=10000?'V3':'V2'};const objects=Object.entries(client).filter(([_,value])=>value&&typeof value==='object');objects.forEach(([toplevelKey,toplevel])=>{const found=Object.entries(toplevel).find(([_,value])=>(value&&typeof value==='object'&&'sitekey' in value&&'size' in value));if(typeof toplevel==='object'&&toplevel instanceof HTMLElement&&toplevel.tagName==='DIV'){data.pageurl=toplevel.baseURI}if(found){const[sublevelKey,sublevel]=found;data.sitekey=sublevel.sitekey;const callbackKey=data.version==='V2'?'callback':'promise-callback';const callback=sublevel[callbackKey];if(!callback){data.callback=null;data.function=null}else{data.function=callback;const keys=[cid,toplevelKey,sublevelKey,callbackKey].map((key)=>`['${key}']`).join('');data.callback=`___grecaptcha_cfg.clients${keys}`}}});return data})[0].callback}return[]} document.getElementsByClassName('LinkContainer-sc-1t58wcv-0 fPyYIP')[0].innerHTML=findRecaptchaClients();"
		_, err := wd.ExecuteScript(discoverV2Clients, nil) // Chrome Inspect > Console > Type '___grecaptcha_cfg.clients[0].' until callback discovered
		if err != nil {
			_, _ = red.Println("Error locating v2 Callback:", err)
		} else {
			// Retrieve the Captcha Callback placed in selected <p> tag via class name
			retrieveCallback := "return document.getElementsByClassName('LinkContainer-sc-1t58wcv-0 fPyYIP')[0].innerHTML"
			jsDiscoveredCallback, err := wd.ExecuteScript(retrieveCallback, nil)
			if err != nil {
				_, _ = red.Println("Error locating v2 Callback:", err)
			} else {
				callbackConv := fmt.Sprintf("%s", jsDiscoveredCallback)
				callbackObj := callbackConv[32 : len(callbackConv)-19]

				_, _ = cyan.Println("v2 Callback Object:", callbackObj)
				genCallback := fmt.Sprintf(".%s.%s.", callbackObj, callbackObj)
				captchaCallback = genCallback // Update empty callback from remote DB with discovered callback object
			}

			// Callback Execution
			_, err = wd.ExecuteScript(fmt.Sprintf("___grecaptcha_cfg.clients[0]%scallback('%s')", captchaCallback, solved), nil) // Chrome Inspect > Console > Type '___grecaptcha_cfg.clients[0].' until callback discovered
			if err != nil {
				_, _ = red.Println("Target Callback: ___grecaptcha_cfg.clients[0]" + captchaCallback + "callback(solvedToken)")
				panic(fmt.Sprintf("[✕](v2) Callback not executed: %s", err)) // ReCaptcha Callback function wasn't executed
			} else {
				log.Println("[✓](v2) ReCaptcha Callback executed:", fmt.Sprintf("___grecaptcha_cfg.clients[0]%scallback('%s')", captchaCallback, solved))

				scrollDownSpotify(wd)                // Justify at page bottom before form submission
				time.Sleep(holdItTime * time.Second) // Wait

				injectSignupBtn(wd)
			}
		}
	}
}

func injectSignupBtn(wd selenium.WebDriver) {
	_, err := wd.ExecuteScript("document.querySelector('.Button-sc-8cs45s-0.jgLyVk').click();", nil) // Chrome Inspect > Console > Type '___grecaptcha_cfg.clients[0].' until callback discovered
	if err != nil {
		_, _ = red.Println("Signup Button click failed:", err)

		// Submit Form via Register button (FALLBACK)
		registerBtn, err := wd.FindElement(selenium.ByXPATH, signupBtn)
		if err != nil {
			_, _ = red.Println("[✕] Signup Button wan't located (FALLBACK):", err)
		}
		if err := registerBtn.Click(); err != nil {
			_, _ = red.Println("[✕] Signup Button click failed (FALLBACK):", err) // ReCaptcha Callback function wasn't executed
			restartSpotify(wd)
		} else {
			log.Println("[✓] Submit button clicked! (FALLBACK)")

			registrationCheckSpotify(wd)
		}
	} else {
		// Submit Form via Register button (Ensures the button was really clicked upon)
		registerBtn, err := wd.FindElement(selenium.ByXPATH, signupBtn)
		if err != nil {
			_, _ = red.Println("[✕] Signup Button wan't located:", err)
		}
		if err := registerBtn.Click(); err != nil {
			_, _ = red.Println("[✕] Signup Button click failed:", err) // ReCaptcha Callback function wasn't executed
			restartSpotify(wd)
		} else {
			log.Println("[✓] Submit button clicked! (x2)")

			time.Sleep(holdItTime * time.Second) // Wait
			registrationCheckSpotify(wd)
		}
	}
}

func scrollDownSpotify(wd selenium.WebDriver) {
	_, err := wd.ExecuteScript("window.scrollBy(0,1500);", nil)
	if err != nil {
		panic(fmt.Sprintf("[✕] Window Scroll Down Error %s", err)) // ReCaptcha Key wasn't submitted back to website.
	} else {
		log.Println("[✓] Scrolled down page!")
	}
}

func scrollUpSpotify(wd selenium.WebDriver) {
	_, err := wd.ExecuteScript("window.scrollBy(0,-1500);", nil)
	if err != nil {
		panic(fmt.Sprintf("[✕] Window Scroll Up Error %s", err)) // ReCaptcha Key wasn't submitted back to website.
	} else {
		log.Println("[✓] Scrolled up page!")
	}
}

func registrationCheckSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("registrationCheckSpotify('Detected for proxy service by Spotify') panic occured:", r)

			time.Sleep(1 * time.Minute) // Delay Wait
			wd.Quit()

			time.Sleep(2 * time.Minute) // Delay Wait
			restartSpotify(wd)
		}
	}()

	scrollUpSpotify(wd)
	time.Sleep(holdItTime * time.Second) // Wait

	_, err := wd.FindElement(selenium.ByXPATH, isSignUpDone)
	if err != nil {
		_, _ = cyan.Println("registrationCheckSpotify('2captcha.com provided valid token')") // , err) but since not shown don't display since ignoring error

		time.Sleep(holdItTime * time.Second) // Wait

		// Double check authentication
		_, err := wd.FindElement(selenium.ByXPATH, accMenuBtn)
		if err != nil {
			panic(err)
		} else {
			_, err := wd.FindElement(selenium.ByXPATH, accMenuBtn)
			if err != nil {
				_, _ = red.Println("registrationCheckSpotify('WebPlayer not visble'):", err)
			} else {
				_, _ = cyan.Println(fmt.Sprintf("\n\tUser successfully registered! [NewCreatedAccount #%d]\n", accountCreationCount))

				accountCreationCount++
				Database.AddSpotifyAccount(UserGenEmailSpotifyGlbl, UserGenPasswordSpotifyGlbl) // ADD credentials to database
				_, _ = magenta.Println("\tNew credentials added to remote DB!")

				accountsInDB = accountsInDB + 1 // Sync DB call

				// Create required number of accounts before streaming again
				if accountCreationCount >= accountsRequired {
					// Check if enough accounts have been created before proceeding
					if accountsInDB >= accountsRequired {
						restartSpotify(wd)
					} else {
						initSpotifyAccountCreation(wd) // Continue to create accounts
					}
				} else {
					initSpotifyAccountCreation(wd) // Continue to create accounts
				}
			}
		}
	} else {
		_, _ = red.Println("registrationCheckSpotify('Signup Visible'): 'Signup with Facebook' located")

		accountCreationErrorCount++

		if accountCreationErrorCount%3 == 0 {
			accountCreationErrorCount = 0

			time.Sleep(2 * time.Minute) // Delay Wait
			restartSpotify(wd)
		} else {
			scrollDownSpotify(wd)
			injectSignupBtn(wd)
		}
	}
}

func misplacedTrackCheckSpotify(wd selenium.WebDriver) {
	// Recover from panic
	defer func() {
		if r := recover(); r != nil {
			_, _ = red.Println("(misplacedTrackCheck) panic occured: ", r)

			wd.Refresh()
			misplacedTrackCheckSpotify(wd)
		}
	}()

	trackTitleDiv, err := wd.FindElement(selenium.ByXPATH, trackTitleDiv)
	if err != nil {
		panic(err)
	}

	trackTitle, _ := trackTitleDiv.Text()
	if trackTitle == "Just Saying(feat.Baby G)" { // Forbid this song from obtaining streams
		skipBtn, err := wd.FindElement(selenium.ByXPATH, skipBtn)
		if err != nil {
			panic(err)
		}
		if err := skipBtn.Click(); err != nil {
			misplacedTrackCheckFail++
			_, _ = red.Printf("\tSkip button click failed.\n")

			if misplacedTrackCheckFail%3 == 0 {
				misplacedTrackCheckFail = 0 // Reset

				restartSpotify(wd)
			}
		} else {
			log.Printf("#(%d) Misplaced track '%s' skipped.", shuffleIterationSpotify, trackTitle)
		}
	}
}

func estimatedRoyaltyCalcSpotify() {
	newEst := estimatedFigCount

	_, _ = green.Println(fmt.Sprintf("\tEstimation of Royalties: ~$%.2f", newEst))
}

func pricePerStreamGenSpotify(x float64, y float64) string {
	// Used to get an estimated average price per stream

	f := randPriceFloatsSpotify(x, y, 1)
	s := fmt.Sprintf("%f", f)            // Convert float64[] to string
	a := strings.Replace(s, "[", "", -1) // Trim [ from string
	b := strings.Replace(a, "]", "", -8) // Trim ] from string
	c, _ := strconv.ParseFloat(b, 64)    // Convert string to float
	d := fmt.Sprintf("%.5f", c)

	estimatedFigCount = estimatedFigCount + c

	return d
}

func randPriceFloatsSpotify(min, max float64, n int) []float64 {
	rand.Seed(time.Now().UTC().UnixNano())

	res := make([]float64, n)

	for i := range res {
		res[i] = min + rand.Float64()*(max-min)
	}

	return res
}
