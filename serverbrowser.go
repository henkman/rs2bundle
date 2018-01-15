package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"encoding/json"

	"github.com/henkman/steamwebapi"
	"github.com/zserge/webview"
)

type Native struct {
	key     string
	dir     string
	Servers []steamwebapi.Server `json:"servers"`
}

func (n *Native) UpdateServers(showempty bool) {
	filter := `\gamedir\RS2`
	if !showempty {
		filter += `\empty\1`
	}
	servers, err := steamwebapi.GetServerList(n.key, 5000, filter)
	if err != nil {
		servers = []steamwebapi.Server{}
	}
	n.Servers = servers
}

func (n *Native) RunShowStats(server string) {
	prog := filepath.Join(n.dir, "serverstats.exe")
	cmd := exec.Command("cmd", "/C", "start", prog, "-s", server)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func errorPopup(msg string) {
	const HTML = `<!doctype html><html><body><h1>%s</h1></body></html>`
	w := webview.New(webview.Settings{
		Title: "error",
		URL: `data:text/html,` +
			url.PathEscape(fmt.Sprintf(HTML, msg)),
	})
	w.Run()
	w.Exit()
}

func main() {
	var config struct {
		Key string `json:"key"`
	}
	{
		fd, err := os.Open("serverbrowser.json")
		if err != nil {
			errorPopup(err.Error())
			return
		}
		err = json.NewDecoder(fd).Decode(&config)
		fd.Close()
		if err != nil {
			errorPopup(err.Error())
			return
		}
	}
	ex, err := os.Executable()
	if err != nil {
		errorPopup(err.Error())
		return
	}
	dir := filepath.Dir(ex)
	url := "file:///" + filepath.ToSlash(dir) + "/r/serverbrowser.html"
	w := webview.New(webview.Settings{
		Title:     "browser",
		URL:       url,
		Width:     1280,
		Height:    680,
		Resizable: true,
	})
	defer w.Exit()
	w.Dispatch(func() {
		w.Bind("native", &Native{key: config.Key, dir: dir})
	})
	w.Run()
}
