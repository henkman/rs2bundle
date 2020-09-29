package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/henkman/steamquery"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func main() {
	if len(os.Args) != 2 {
		return
	}
	addr, err := net.ResolveUDPAddr("udp", os.Args[1])
	if err != nil {
		return
	}

	model := new(ServerModel)

	go func() {
		model.Refresh(addr)
		model.Sort(model.SortedColumn(), model.SortOrder())
		model.PublishRowsReset()
	}()

	const (
		TABLE_WIDTH  = 440
		TABLE_HEIGHT = 600
	)
	var tv *walk.TableView
	var mw *walk.MainWindow
	var appIcon, _ = walk.NewIconFromResourceId(3)
	if err := (MainWindow{
		Icon:     appIcon,
		AssignTo: &mw,
		Title:    "Stats",
		Size: Size{
			Width:  TABLE_WIDTH,
			Height: TABLE_HEIGHT,
		},
		Layout: VBox{
			MarginsZero: true,
		},
		Children: []Widget{
			Composite{
				Layout: Grid{
					Columns: 4,
					Margins: Margins{
						Left:   5,
						Top:    5,
						Right:  0,
						Bottom: 0,
					},
				},
				Children: []Widget{
					PushButton{
						Text: "Refresh",
						OnClicked: func() {
							model.Refresh(addr)
							model.Sort(model.SortedColumn(), model.SortOrder())
							model.PublishRowsReset()
						},
						MaxSize: Size{80, 30},
					},
					PushButton{
						Text: "Start",
						OnClicked: func() {
							url := fmt.Sprintf("steam://rungameid/%d/",
								model.gameID)
							cmd := exec.Command("cmd", "/C", "start", url)
							cmd.Run()
						},
						MaxSize: Size{80, 30},
					},
					PushButton{
						Text: "Steam",
						OnClicked: func() {
							cmd := exec.Command("cmd", "/C", "start",
								"steam://connect/"+addr.String())
							cmd.Start()
						},
						MaxSize: Size{80, 30},
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
				},
			},
			Composite{
				Alignment: AlignHNearVCenter,
				Layout: Grid{
					Columns: 2,
					Margins: Margins{
						Left:   5,
						Top:    0,
						Right:  0,
						Bottom: 0,
					},
				},
				Children: []Widget{
					Label{
						Text:    "Name:",
						MaxSize: Size{45, 20},
						Font:    Font{Family: "Segoe UI", PointSize: 12},
					},
					Label{
						AssignTo: &model.nameLabel,
						MinSize:  Size{380, 20},
						Font:     Font{Family: "Segoe UI", PointSize: 12, Bold: true},
					},
				},
			},
			Composite{
				AlwaysConsumeSpace: true,
				Alignment:          AlignHNearVCenter,
				Layout: Grid{
					Columns: 8,
					Margins: Margins{
						Left:   5,
						Top:    0,
						Right:  0,
						Bottom: 0,
					},
				},
				Children: []Widget{
					Label{
						Text:    "Map:",
						MaxSize: Size{35, 20},
						Font:    Font{Family: "Segoe UI", PointSize: 11},
					},
					Label{
						AssignTo: &model.mapLabel,
						MinSize:  Size{180, 20},
						Font:     Font{Family: "Segoe UI", PointSize: 11, Bold: true},
					},
					HSpacer{Size: 5},
					Label{
						Text:    "Players:",
						MaxSize: Size{50, 20},
						Font:    Font{Family: "Segoe UI", PointSize: 11},
					},
					Label{
						AssignTo: &model.playersLabel,
						MaxSize:  Size{40, 20},
						Font:     Font{Family: "Segoe UI", PointSize: 11},
					},
					HSpacer{Size: 5},
					Label{
						Text:    "Ping:",
						MaxSize: Size{35, 20},
						Font:    Font{Family: "Segoe UI", PointSize: 11},
					},
					Label{
						AssignTo: &model.pingLabel,
						MaxSize:  Size{55, 20},
						Font:     Font{Family: "Segoe UI", PointSize: 11},
					},
				},
			},
			TableView{
				AssignTo: &tv,
				MinSize: Size{
					Width:  TABLE_WIDTH,
					Height: TABLE_HEIGHT,
				},
				AlwaysConsumeSpace: true,
				MultiSelection:     true,
				AlternatingRowBG:   true,
				ColumnsOrderable:   true,
				CustomRowHeight:    24,
				Columns: []TableViewColumn{
					{Title: "Name", Width: 160},
					{Title: "Score", Width: 50},
					{Title: "Online", Width: 120},
					{Title: "Score/Second", Width: 90},
				},
				Model: model,
				OnKeyPress: func(key walk.Key) {
					m := walk.ModifiersDown()
					if m&walk.ModControl == walk.ModControl {
						if key == walk.KeyC {
							var sb strings.Builder
							idxs := tv.SelectedIndexes()
							sort.Ints(idxs)
							for _, i := range idxs {
								player := model.players[model.shown[i]]
								fmt.Fprintln(&sb, player)
							}
							walk.Clipboard().SetText(sb.String())
						} else if key == walk.KeyA {
							tv.SetSelectedIndexes(model.shown)
						}
					}
				},
				OnItemActivated: func() {
					player := model.players[model.shown[tv.CurrentIndex()]]
					u := fmt.Sprint("https://steamcommunity.com/search/users/#text=", url.QueryEscape(player.Name))
					exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	r := mw.Bounds()
	scrWidth := int(win.GetSystemMetrics(win.SM_CXSCREEN))
	scrHeight := int(win.GetSystemMetrics(win.SM_CYSCREEN))
	mw.SetBounds(walk.Rectangle{
		X:      int((scrWidth - r.Width) / 2),
		Y:      int((scrHeight - r.Height) / 2),
		Width:  r.Width,
		Height: r.Height,
	})
	mw.Run()
}

type Column = int

const (
	ColumnName Column = iota
	ColumnScore
	ColumnOnline
	ColumnScorePerSecond
)

type Player struct {
	Name           string
	Score          int32
	Online         time.Duration
	ScorePerSecond float64
}

type ServerModel struct {
	walk.TableModelBase
	walk.SorterBase
	filterTextEdit *walk.TextEdit
	nameLabel      *walk.Label
	mapLabel       *walk.Label
	playersLabel   *walk.Label
	pingLabel      *walk.Label
	players        []Player
	shown          []int
	gameID         uint64
}

func (m *ServerModel) RowCount() int {
	return len(m.shown)
}

func (m *ServerModel) Value(row, col int) interface{} {
	player := m.players[m.shown[row]]
	switch col {
	case ColumnName:
		return player.Name
	case ColumnScore:
		return player.Score
	case ColumnOnline:
		return player.Online
	case ColumnScorePerSecond:
		return fmt.Sprintf("%.6f", player.ScorePerSecond)
	}
	panic("unexpected col")
}

func (m *ServerModel) Sort(col int, order walk.SortOrder) error {
	sort.SliceStable(m.shown, func(i, j int) bool {
		a, b := m.players[m.shown[i]], m.players[m.shown[j]]
		switch col {
		case ColumnName:
			if order == walk.SortAscending {
				return a.Name < b.Name
			}
			return a.Name > b.Name
		case ColumnScore:
			if order == walk.SortAscending {
				return a.Score < b.Score
			}
			return a.Score > b.Score
		case ColumnOnline:
			if order == walk.SortAscending {
				return a.Online < b.Online
			}
			return a.Online > b.Online
		case ColumnScorePerSecond:
			if order == walk.SortAscending {
				return a.ScorePerSecond < b.ScorePerSecond
			}
			return a.ScorePerSecond > b.ScorePerSecond
		}
		panic("unreachable")
	})
	return m.SorterBase.Sort(col, order)
}

func (m *ServerModel) Refresh(addr *net.UDPAddr) {
	players, _, err := steamquery.QueryPlayers(addr)
	if err != nil {
		return
	}
	m.shown = m.shown[:0]
	m.players = m.players[:0]
	filter := strings.ToLower(m.filterTextEdit.Text())
	for i, player := range players {
		pls := Player{
			Name:           strings.TrimSpace(player.Name),
			Score:          player.Score,
			Online:         player.Duration,
			ScorePerSecond: float64(player.Score) / float64(player.Duration/time.Second),
		}
		m.players = append(m.players, pls)
		if filter == "" ||
			strings.Contains(strings.ToLower(pls.Name), filter) {
			m.shown = append(m.shown, i)
		}
	}
	info, ping, err := steamquery.QueryInfo(addr)
	if err != nil {
		return
	}
	name := strings.TrimSpace(info.Name)
	m.nameLabel.SetText(name)
	m.nameLabel.SetToolTipText(name)
	m.mapLabel.SetText(info.Map)
	m.mapLabel.SetToolTipText(info.Map)
	m.playersLabel.SetText(fmt.Sprintf("%d/%d", len(players), info.MaxPlayers))
	m.pingLabel.SetText(ping.Truncate(time.Millisecond).String())
	m.gameID = info.GameID
}

func (m *ServerModel) Filter() {
	filter := m.filterTextEdit.Text()
	if filter == "" {
		m.shown = m.shown[:0]
		for i, _ := range m.players {
			m.shown = append(m.shown, i)
		}
		return
	}
	filter = strings.ToLower(filter)
	m.shown = m.shown[:0]
	for i, player := range m.players {
		if strings.Contains(strings.ToLower(player.Name), filter) {
			m.shown = append(m.shown, i)
		}
	}
}
