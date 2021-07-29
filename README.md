# Moonlandr

It's more than just a Record Label.

> Backend, written in Go. (Originally C# circa July 2019)
> - [Selenoid](https://github.com/aerokube/selenoid) (Selenium Browser Automation)
> - [MySQL](github.com/go-sql-driver/mysql) remote database connection. (go-mysql)
> - [2Captcha](https://2captcha.com/) Forked API. (ReCaptcha v2 + v3 Bypass)
> - [Kubernetes](https://github.com/kubernetes/kubernetes). (Cluster Infrastructure)
> - [DigitalOcean](https://www.digitalocean.com/). (Production MySQL & Kubernetes Cloud Hosting Provider)

### Spotify: Worker's Job
- Check remote DB for any available Spotify accounts, if none create some.
- Once created and ReCaptcha bypassed, add account to remote DB. (ReCaptcha Bypass Embedded via 2captcha.com)
- Randomly selected Spotify account from remote DB, signin account via Selenoid. (Selenium)
- Load Target Artist Profile, start the artist/client's discography.
- Wait 45 seconds to 1 minute before initializing streaming loop. (30 seconds considered a stream)
- After 45 seconds to 1 minute, Selenium shuffles to next song. (loop)
- Self-sufficient error handlers if stream loop is interrupted. (Session Timeout/Advertisements)
- Repeat once a desired amount of stream iterations have been reached. (For example, every 1,200 listens)

## Local Deployment
```
go run main.go
```

Moonlandr supports flags, deploy a supported service:
```
go run main.go --deploy spotify
```