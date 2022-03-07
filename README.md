<h1 align="center">updatebot ✏️</h1>

A simple CLI util for change your Discord bot's username and avatar while this functionality is disabled on the dev portal.

### Usage

Note: token is first read from `$DISCORD_TOKEN`. If no such variable exists, the program will prompt you to input the token.

**Running with Docker (easiest)** `docker run -t benricheson101/updatebot [--username=""] [--avatar=""]`

**Download executable** Download the corresponding executable from the [releases page](https://github.com/benricheson101/updatebot/releases/latest) and run it from the terminal.

**Build from source**
```sh
$ git clone git@github.com:benricheson101/updatebot.git
$ cd updatebot
$ go build -o updatebot cmd/updatebot/main.go
$ ./updatebot
```

### Command-Line Flags
```
Usage of updatebot:
  -avatar url or file path
    	the new avatar for the bot. either a url or file path
  -username string
    	the new username for the bot
```
