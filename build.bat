@go get -u -v
@go build -ldflags="-s -w -H windowsgui" serverbrowser.go
@go build -ldflags="-s -w -H windowsgui" serverstats.go
@go build -ldflags="-s -w" autojoiner.go