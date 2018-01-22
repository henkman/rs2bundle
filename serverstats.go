package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/henkman/steamquery"
	"github.com/zserge/webview"
)

type Native struct {
	w       webview.WebView
	address *net.UDPAddr
	dir     string
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

func (n *Native) GetServerInfo(cb string) {
	go func() {
		var info struct {
			Name           string              `json:"name"`
			Map            string              `json:"map"`
			PlayerCount    int                 `json:"playerCount"`
			MaxPlayerCount int                 `json:"maxPlayerCount"`
			Ping           time.Duration       `json:"ping"`
			Players        []steamquery.Player `json:"players"`
			Error          string              `json:"error,omitempty"`
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			players, ping, err := steamquery.QueryPlayers(n.address)
			if err != nil {
				info.Error = err.Error()
				wg.Done()
				return
			}
			info.Ping = ping
			info.Players = players
			wg.Done()
		}()
		go func() {
			rules, _, err := steamquery.QueryRules(n.address)
			if err != nil {
				info.Error = err.Error()
				wg.Done()
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
						info.Error = err.Error()
						wg.Done()
						return
					}
					numOpenPublicConnections = tmp
				} else if r.Name == "NumPublicConnections" {
					tmp, err := strconv.Atoi(r.Value)
					if err != nil {
						info.Error = err.Error()
						wg.Done()
						return
					}
					numPublicConnections = tmp
				} else if r.Name == "p2" {
					p2 = r.Value
				}
			}
			info.Name = owningPlayerName
			info.Map = p2
			info.PlayerCount = numPublicConnections - numOpenPublicConnections
			info.MaxPlayerCount = numPublicConnections
			wg.Done()
		}()
		wg.Wait()
		raw, err := json.Marshal(info)
		if err != nil {
			info.Error = err.Error()
			return
		}
		raw = bytes.Replace(raw, []byte(`\`), []byte(`\\`), -1)
		raw = bytes.Replace(raw, []byte(`"`), []byte(`\"`), -1)
		code := fmt.Sprintf(`(function(x) {%s})("%s");`, cb, string(raw))
		n.w.Dispatch(func() {
			n.w.Eval(code)
		})
	}()
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
		w.Bind("native", &Native{w: w, address: addr, dir: dir})
	})
	w.Run()
}
