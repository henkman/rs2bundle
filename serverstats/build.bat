rsrc -ico app.ico -manifest app.manifest -o rsrc.syso
go get -v -d
go build -ldflags="-H windowsgui -s -w"