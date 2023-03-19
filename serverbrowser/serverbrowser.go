package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/henkman/steamwebapi"
	"github.com/ip2location/ip2location-go"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(exe)
	model := new(ServerModel)
	{
		fd, err := os.Open(filepath.Join(dir, "serverbrowser.json"))
		if err != nil {
			panic(err)
		}
		err = json.NewDecoder(fd).Decode(&model.key)
		fd.Close()
		if err != nil {
			panic(err)
		}
	}
	{
		db, err := ip2location.OpenDB(
			filepath.Join(dir, `IP2LOCATION-LITE-DB1.BIN`))
		if err != nil {
			panic(err)
		}
		model.ip2location = db
	}
	cfr, err := makeCountryFlagRenderer(
		filepath.Join(dir, "flags.png"),
		filepath.Join(dir, "flags.csv"))
	if err != nil {
		panic(err)
	}
	normalFont, _ := walk.NewFont("Segoe UI", 9, 0)
	boldFont, _ := walk.NewFont("Segoe UI", 9, walk.FontBold)
	var tv *walk.TableView
	const (
		WINDOW_WIDTH  = 970
		WINDOW_HEIGHT = 632
	)
	var mw *walk.MainWindow

	var appIcon, _ = walk.NewIconFromResourceId(2)
	if err := (MainWindow{
		Icon:     appIcon,
		AssignTo: &mw,
		Title:    "Serverbrowser",
		Size:     Size{WINDOW_WIDTH, WINDOW_HEIGHT},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: Grid{
					Columns: 5,
				},
				MaxSize: Size{510, 32},
				Children: []Widget{
					PushButton{
						Text: "Refresh",
						OnClicked: func() {
							game := games[model.gameComboBox.CurrentIndex()]
							model.Refresh(game, model.showEmptyCheckbox.Checked())
							model.Filter()
							model.Sort(model.SortedColumn(), model.SortOrder())
							model.PublishRowsReset()
						},
						MaxSize: Size{120, 30},
					},
					CheckBox{
						AssignTo: &model.showEmptyCheckbox,
						Text:     "Show empty",
						MaxSize:  Size{120, 30},
						OnCheckedChanged: func() {
							game := games[model.gameComboBox.CurrentIndex()]
							model.Refresh(game, model.showEmptyCheckbox.Checked())
							model.Sort(model.SortedColumn(), model.SortOrder())
							model.PublishRowsReset()
						},
					},
					Label{
						AssignTo: &model.totalPlayerLabel,
						MaxSize:  Size{120, 30},
					},
					TextEdit{
						AssignTo:    &model.filterTextEdit,
						ToolTipText: "Filter",
						MaxSize:     Size{120, 30},
						OnTextChanged: func() {
							model.Filter()
							model.Sort(model.SortedColumn(), model.SortOrder())
							model.PublishRowsReset()
						},
					},
					ComboBox{
						AssignTo: &model.gameComboBox,
						Model:    games,
						MaxSize:  Size{50, 30},
						OnCurrentIndexChanged: func() {
							game := games[model.gameComboBox.CurrentIndex()]
							model.Refresh(game, model.showEmptyCheckbox.Checked())
							model.Sort(model.SortedColumn(), model.SortOrder())
							model.PublishRowsReset()
						},
					},
				},
			},
			TableView{
				MinSize: Size{
					Width:  970,
					Height: 600,
				},
				AssignTo:         &tv,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				MultiSelection:   true,
				CustomRowHeight:  24,
				Columns: []TableViewColumn{
					{Title: "Name", Width: 440},
					{Title: "Map", Width: 140},
					{Title: "Players", Width: 55},
					{Title: "Steamid", Width: 120},
					{Title: "Addr", Width: 135},
					{Title: "Country", Width: 55},
				},
				StyleCell: func(style *walk.CellStyle) {
					server := model.servers[model.shown[style.Row()]]
					switch style.Col() {
					case ColumnName:
						style.Font = boldFont
					case ColumnCountry:
						if canvas := style.Canvas(); canvas != nil {
							bounds := style.Bounds()
							p := walk.Point{bounds.X + 12, bounds.Y}
							cfr.Render(canvas, p, server.Country)
						}
					default:
						style.Font = normalFont
					}
				},
				Model: model,
				OnItemActivated: func() {
					server := model.servers[model.shown[tv.CurrentIndex()]]
					prog := filepath.Join(dir, "serverstats.exe")
					cmd := exec.Command(prog, server.Address)
					cmd.Start()
				},
				OnKeyPress: func(key walk.Key) {
					m := walk.ModifiersDown()
					if m&walk.ModControl == walk.ModControl {
						if key == walk.KeyC {
							var sb strings.Builder
							idxs := tv.SelectedIndexes()
							sort.Ints(idxs)
							for _, i := range idxs {
								server := model.servers[model.shown[i]]
								fmt.Fprintln(&sb, server)
							}
							walk.Clipboard().SetText(sb.String())
						} else if key == walk.KeyA {
							tv.SetSelectedIndexes(model.shown)
						}
					}
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	go func() {
		model.Sort(ColumnPlayers, walk.SortDescending)
		model.gameComboBox.SetCurrentIndex(0) // RS2
	}()

	scrWidth := win.GetSystemMetrics(win.SM_CXSCREEN)
	scrHeight := win.GetSystemMetrics(win.SM_CYSCREEN)
	mw.SetBounds(walk.Rectangle{
		X:      int((scrWidth - WINDOW_WIDTH) / 2),
		Y:      int((scrHeight - WINDOW_HEIGHT) / 2),
		Width:  WINDOW_WIDTH,
		Height: WINDOW_HEIGHT,
	})
	mw.Run()
}

var (
	games = []string{
		"RS2",
		"RO2",
	}
)

type Column = int

const (
	ColumnName Column = iota
	ColumnMap
	ColumnPlayers
	ColumnSteamid
	ColumnAddress
	ColumnCountry
)

type Server struct {
	Name       string
	Map        string
	Players    int
	MaxPlayers int
	Steamid    string
	Address    string
	Country    string
}

type ServerModel struct {
	walk.TableModelBase
	walk.SorterBase
	showEmptyCheckbox *walk.CheckBox
	totalPlayerLabel  *walk.Label
	filterTextEdit    *walk.TextEdit
	gameComboBox      *walk.ComboBox
	servers           []Server
	shown             []int
	ip2location       *ip2location.DB
	key               string
}

func (m *ServerModel) RowCount() int {
	return len(m.shown)
}

func (m *ServerModel) Value(row, col int) interface{} {
	server := m.servers[m.shown[row]]
	switch col {
	case ColumnName:
		return server.Name
	case ColumnMap:
		return server.Map
	case ColumnPlayers:
		return fmt.Sprintf("%d/%d", server.Players, server.MaxPlayers)
	case ColumnSteamid:
		return server.Steamid
	case ColumnAddress:
		return server.Address
	case ColumnCountry:
		return server.Country
	}
	panic("unexpected col")
}

func (m *ServerModel) Sort(col int, order walk.SortOrder) error {
	sort.SliceStable(m.shown, func(i, j int) bool {
		a, b := m.servers[m.shown[i]], m.servers[m.shown[j]]
		switch col {
		case ColumnName:
			if order == walk.SortAscending {
				return a.Name < b.Name
			}
			return a.Name > b.Name
		case ColumnMap:
			if order == walk.SortAscending {
				return a.Map < b.Map
			}
			return a.Map > b.Map
		case ColumnPlayers:
			if order == walk.SortAscending {
				return a.Players < b.Players
			}
			return a.Players > b.Players
		case ColumnSteamid:
			if order == walk.SortAscending {
				return a.Steamid < b.Steamid
			}
			return a.Steamid > b.Steamid
		case ColumnAddress:
			if order == walk.SortAscending {
				return a.Address < b.Address
			}
			return a.Address > b.Address
		case ColumnCountry:
			if order == walk.SortAscending {
				return a.Country < b.Country
			}
			return a.Country > b.Country
		}
		panic("unreachable")
	})
	return m.SorterBase.Sort(col, order)
}

func (m *ServerModel) Refresh(game string, showEmpty bool) {
	f := `\gamedir\` + game
	if !showEmpty {
		f += `\empty\1`
	}
	servers, err := steamwebapi.GetServerList(m.key, 5000, f)
	if err != nil {
		panic(err)
	}
	total := 0
	m.shown = m.shown[:0]
	m.servers = m.servers[:0]
	filter := strings.ToLower(m.filterTextEdit.Text())
	for i, server := range servers {
		s := strings.Split(server.Addr, ":")
		country := "un"
		if len(s) > 0 {
			location, err := m.ip2location.Get_country_short(s[0])
			if err == nil {
				country = strings.ToLower(location.Country_short)
			}
		}
		serv := Server{
			Name:       strings.TrimSpace(server.Name),
			Map:        server.Map,
			Players:    server.Players,
			MaxPlayers: server.MaxPlayers,
			Steamid:    server.Steamid,
			Address:    server.Addr,
			Country:    country,
		}
		m.servers = append(m.servers, serv)
		total += serv.Players
		if filter == "" ||
			strings.Contains(strings.ToLower(serv.Name), filter) ||
			strings.Contains(strings.ToLower(serv.Map), filter) {
			m.shown = append(m.shown, i)
		}
	}
	m.totalPlayerLabel.SetText(fmt.Sprint(total, " players"))
}

func (m *ServerModel) Filter() {
	filter := m.filterTextEdit.Text()
	if filter == "" {
		m.shown = m.shown[:0]
		for i, _ := range m.servers {
			m.shown = append(m.shown, i)
		}
		return
	}
	filter = strings.ToLower(filter)
	m.shown = m.shown[:0]
	for i, server := range m.servers {
		if strings.Contains(strings.ToLower(server.Name), filter) ||
			strings.Contains(strings.ToLower(server.Map), filter) {
			m.shown = append(m.shown, i)
		}
	}
}

type CountryFlagRenderer struct {
	bitmap  *walk.Bitmap
	offsets map[string]walk.Point
}

func makeCountryFlagRenderer(imagePath, offsetsPath string) (CountryFlagRenderer, error) {
	var cfr CountryFlagRenderer
	bmp, err := walk.NewBitmapFromFileForDPI(imagePath, 96)
	if err != nil {
		return cfr, err
	}
	offsets := map[string]walk.Point{}
	{
		fd, err := os.Open(offsetsPath)
		if err != nil {
			return cfr, err
		}
		bin := bufio.NewReader(fd)
		for {
			var country string
			var x, y int
			_, err := fmt.Fscanf(bin, "%s %d %d\n", &country, &x, &y)
			if err != nil {
				break
			}
			offsets[country] = walk.Point{x, y}
		}
		fd.Close()
	}
	cfr = CountryFlagRenderer{
		bitmap:  bmp,
		offsets: offsets,
	}
	return cfr, nil
}

func (cfr *CountryFlagRenderer) Render(c *walk.Canvas, p walk.Point, code string) error {
	o, ok := cfr.offsets[code]
	if !ok {
		return errors.New("country " + code + " unknown")
	}
	const (
		W = 32
		H = 20
	)
	src := walk.Rectangle{o.X, o.Y + 6, W, H}
	dst := walk.Rectangle{p.X, p.Y, W, H}
	return c.DrawBitmapPart(cfr.bitmap, dst, src)
}
