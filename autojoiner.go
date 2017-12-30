package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AllenDang/w32"
	"github.com/henkman/steamquery"
	"github.com/skratchdot/open-golang/open"
)

func TypeKeycode(keycode uint16, shift bool) {
	const (
		KEYEVENTF_KEYUP    = 0x0002
		KEYEVENTF_SCANCODE = 0x0008
	)
	var in w32.INPUT
	in.Type = w32.INPUT_KEYBOARD
	in.Ki.WVk = keycode
	var send []w32.INPUT
	if shift {
		var shiftin w32.INPUT
		shiftin.Type = w32.INPUT_KEYBOARD
		shiftin.Ki.WVk = w32.VK_SHIFT
		send = []w32.INPUT{shiftin, in}
	} else {
		send = []w32.INPUT{in}
	}
	for i, _ := range send {
		send[i].Ki.DwFlags = 0
	}
	w32.SendInput(send)
	for i, _ := range send {
		send[i].Ki.DwFlags = 2
	}
	w32.SendInput(send)
}

func TypeText(text string) {
	for _, c := range text {
		if c >= 'a' && c <= 'z' {
			TypeKeycode(uint16(c-'a'+'A'), false)
		} else if c == '.' {
			TypeKeycode(w32.VK_OEM_PERIOD, false)
		} else if c == ':' {
			TypeKeycode(w32.VK_OEM_PERIOD, true)
		} else {
			TypeKeycode(uint16(c), false)
		}
	}
}

func GetGameWindow() w32.HWND {
	return w32.FindWindowExW(
		w32.HWND(0),
		w32.HWND(0),
		nil,
		syscall.StringToUTF16Ptr("Rising Storm 2: Vietnam"),
	)
}

func ConnectServer(server string, gamewindow w32.HWND) {
	w32.ShowWindow(gamewindow, w32.SW_RESTORE)
	w32.SetForegroundWindow(gamewindow)
	TypeKeycode(w32.VK_F3, false)
	TypeText("open " + server)
	TypeKeycode(w32.VK_RETURN, false)
}

func RunGame(startupseconds uint) {
	open.Run("steam://run/418460/")
	time.Sleep(time.Second * time.Duration(startupseconds))
}

func main() {
	var opts struct {
		Server           string
		Sleepseconds   uint
		Startupseconds uint
		Nonsupremacy   bool
	}
	flag.StringVar(&opts.Server, "s", "", "server")
	flag.UintVar(&opts.Sleepseconds, "st", 3, "sleep in seconds between tries")
	flag.UintVar(&opts.Sleepseconds, "sg", 25, "startup time of game")
	flag.BoolVar(&opts.Nonsupremacy, "ns", false, "no supremacy")
	flag.Parse()
	if opts.Server == "" {
		flag.Usage()
		return
	}
	addr, err := net.ResolveUDPAddr("udp", opts.Server)
	if err != nil {
		log.Fatal(err)
	}
	info, err := steamquery.QueryInfo(addr)
	if err != nil {
		log.Fatal(err)
	}
	if GetGameWindow() == w32.HWND(0) {
		RunGame(opts.Startupseconds)
	}
	for {
		rs, err := steamquery.QueryRules(addr)
		if err != nil {
			log.Fatal(err)
		}
		var vals struct {
			Map       string
			OpenSpots int
			MaxSpots  int
		}
		for _, r := range rs {
			if r.Name == "NumOpenPublicConnections" {
				vals.OpenSpots, _ = strconv.Atoi(r.Value)
			} else if r.Name == "NumPublicConnections" {
				vals.MaxSpots, _ = strconv.Atoi(r.Value)
			} else if r.Name == "p2" {
				vals.Map = r.Value
			}
		}
		if opts.Nonsupremacy && strings.HasPrefix(vals.Map, "VNSU-") {
			fmt.Printf(
				"%s: current map type is supremacy (players: %d, max: %d, map: %s)\n",
				info.Name, vals.MaxSpots-vals.OpenSpots, vals.MaxSpots, vals.Map)
			time.Sleep(time.Second * time.Duration(opts.Sleepseconds))
		} else if vals.OpenSpots <= 0 {
			fmt.Printf("%s: still full (map: %s)\n",
				info.Name, vals.Map)
			time.Sleep(time.Second * time.Duration(opts.Sleepseconds))
		} else {
			host := fmt.Sprintf("%s:%d", addr.IP.String(), info.Port)
			fmt.Printf("connecting to %s %s (%d players)\n",
				info.Name, host, vals.MaxSpots-vals.OpenSpots)
			hwnd := GetGameWindow()
			if hwnd == w32.HWND(0) {
				RunGame(opts.Startupseconds)
				hwnd = GetGameWindow()
			}
			ConnectServer(host, hwnd)
			break
		}
	}
}
