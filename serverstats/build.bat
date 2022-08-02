rsrc -ico app.ico -manifest app.manifest -o rsrc.syso
go get -v -d
go build -trimpath -tags walk_use_cgo -ldflags="-H windowsgui -s -w"