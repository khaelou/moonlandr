# Moonlandr

It's more than just a Record Label.

> Backend, written in Go. (Originally C# in July 2019)
> - Selenoid (Selenium Browser Automation)
> - Remote DB connection. (go-sql)
> - 2Captcha API. (ReCaptcha v2 + v3 Solving)
> - Kubernetes. (Production  Cluster Deployment)

## Behind the Scenes
- Check remote DB for any available spotify accounts, if none create some accounts.
- Creates User Accounts via Spotify, adds accounts to remote DB. (ReCaptcha Bypass Embedded)
- Random Spotify Account retrieved from remote DB, log in using through Selenoid. (Selenium)
- Load Target Artist Profile, start discography.
- Wait 45 seconds to 1 minute before initializing streaming loop.
- After 30 seconds - 1 minute, Selenium shuffles to next song. (loop)
- Log account out or quit Selenium session if streaming loop is interrupted. (Session Timeout/Advertisements)
- Repeat after a set amount of stream iterations. (For example, every 5,000 listens)
