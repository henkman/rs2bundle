# rs2bundle
tools for Rising Storm 2: Vietnam to show stats and browse available servers

[**Download here**](https://github.com/henkman/rs2bundle/releases/download/1.3/rs2bundle.zip) remember to configure server browser (see below)

![screenshot](https://github.com/henkman/rs2bundle/raw/master/screenshot.jpg "Screenshot")

## tools
- serverbrowser.exe shows all available (populated) servers
- serverstats.exe shows stats of one server

## building
1. [install go](https://golang.org/doc/install)
2. run 'serverbrowser/build.bat'
3. run 'serverstats/build.bat'

## remove startup movies for faster start
set launch options in steam to `-nostartupmovies`

# server browser configuration
- create file serverbrowser.json
- put {key:"$yourkey"} into it
- $yourkey is [the key from steam](https://steamcommunity.com/dev/apikey)
