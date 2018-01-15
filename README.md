# rs2bundle
tools for Rising Storm 2: Vietnam to show stats, browse available servers and auto join servers

[**Download here**](https://github.com/henkman/rs2bundle/releases/download/1.0/rs2bundle.zip)

![screenshot](https://github.com/henkman/rs2bundle/raw/master/screenshot.jpg "Screenshot")

## tools
- serverbrowser.exe shows all available populated servers
- serverstats.exe shows stats of one server
- autojoiner.exe waits until a place is free on a server and then joins you in

## building
1. [install go](https://golang.org/doc/install)
2. run 'go get -u -v github.com/henkman/rs2bundle'
3. run 'go get -ldflags="-s -w" github.com/tdewolff/minify/cmd/minify'
4. run 'build.bat'

## remove startup movies for faster start
set launch options in steam to `-nostartupmovies`

# server browser configuration
- create file serverbrowser.json
- put {"key": "$yourkey"} into it
- $yourkey is [the key from steam](https://steamcommunity.com/dev/apikey)
