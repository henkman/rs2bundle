rsrc -manifest serverbrowser.manifest -o rsrc.syso
go get -v -d
go build -ldflags="-H windowsgui -s -w"