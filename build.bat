@go build -ldflags="-s -w -H windowsgui" serverbrowser.go
@go build -ldflags="-s -w -H windowsgui" serverstats.go
@go build -ldflags="-s -w" autojoiner.go
@minify html/serverbrowser.html -o r/serverbrowser.html
@minify html/serverstats.html -o r/serverstats.html