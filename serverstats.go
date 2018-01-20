package main

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/henkman/steamquery"
	"github.com/zserge/webview"
)

type Info struct {
	Name       string
	Map        string
	Players    int
	MaxPlayers int
}

type Native struct {
	address *net.UDPAddr
	dir     string
	Info    Info                `json:"info"`
	Players []steamquery.Player `json:"players"`
	Ping    time.Duration       `json:"ping"`
}

func (n *Native) RunJoin() {
	prog := filepath.Join(n.dir, "autojoiner.exe")
	cmd := exec.Command("cmd", "/C", "start", prog, "-s",
		n.address.String())
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func (n *Native) RunSteamJoin() {
	cmd := exec.Command("cmd", "/C", "start",
		"steam://connect/"+n.address.String())
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func (n *Native) UpdatePlayers() {
	players, ping, err := steamquery.QueryPlayers(n.address)
	if err != nil {
		n.Players = []steamquery.Player{}
		return
	}
	n.Ping = ping
	n.Players = players
}

func (n *Native) UpdateInfo() {
	rules, _, err := steamquery.QueryRules(n.address)
	if err != nil {
		n.Info = Info{}
		return
	}
	var owningPlayerName string
	var p2 string
	var numOpenPublicConnections int
	var numPublicConnections int
	for _, r := range rules {
		if r.Name == "OwningPlayerName" {
			owningPlayerName = r.Value
		} else if r.Name == "NumOpenPublicConnections" {
			tmp, err := strconv.Atoi(r.Value)
			if err != nil {
				n.Info = Info{}
				return
			}
			numOpenPublicConnections = tmp
		} else if r.Name == "NumPublicConnections" {
			tmp, err := strconv.Atoi(r.Value)
			if err != nil {
				n.Info = Info{}
				return
			}
			numPublicConnections = tmp
		} else if r.Name == "p2" {
			p2 = r.Value
		}
	}
	n.Info = Info{
		Name:       owningPlayerName,
		Map:        p2,
		Players:    numPublicConnections - numOpenPublicConnections,
		MaxPlayers: numPublicConnections,
	}
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
	var opts struct {
		Server string
	}
	flag.StringVar(&opts.Server, "s", "", "server")
	flag.Parse()
	if opts.Server == "" {
		errorPopup("specify server using -s parameter in form ip:port")
		return
	}
	ex, err := os.Executable()
	if err != nil {
		errorPopup(err.Error())
		return
	}
	addr, err := net.ResolveUDPAddr("udp", opts.Server)
	if err != nil {
		errorPopup(err.Error())
		return
	}
	dir := filepath.Dir(ex)
	url := "file:///" + filepath.ToSlash(dir) + "/r/serverstats.html"
	w := webview.New(webview.Settings{
		Title:     "stats",
		URL:       url,
		Width:     680,
		Height:    715,
		Resizable: true,
	})
	defer w.Exit()
	w.Dispatch(func() {
		w.Bind("native", &Native{address: addr, dir: dir})
	})
	w.Run()
}
