package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	HEADING_SIZE = 10
	TITLE        = `  _____                                _____                      ___  __   _ ___   
 / ___/__  ___ _    _____ ___ _____   / ___/__ ___ _  ___   ___  / _/ / /  (_) _/__ 
/ /__/ _ \/ _ \ |/|/ / _  / // (_-<  / (_ / _ /  ' \/ -_)  / _ \/ _/ / /__/ / _/ -_)
\___/\___/_//_/__,__/\_,_/\_, /___/  \___/\_,_/_/_/_/\__/  \___/_/  /____/_/_/ \__/ 
                         /___/                                                     
`
)

// Style
var (
	//Map view settings
	cyan   = lipgloss.Color("86")
	purple = lipgloss.Color("201")
	orange = lipgloss.Color("202")
	colors = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(cyan),
		lipgloss.NewStyle().Foreground(purple),
		lipgloss.NewStyle().Foreground(orange),
	}

	//Map preset settings
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(purple)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
)

const (
	// Game State
	Mapping        = 0
	Playing        = 1
	PresetChoosing = 2
	// Edit State
	Observing = 0
	Removing  = 1
	Adding    = 2
	// Preset choices
	RAND     = "Random Fill"
	EDGES    = "Edge tracing"
	PILLARS  = "Pillars"
	ROWS     = "Rows"
	DOTTED   = "Dotted Lines"
	THREADS  = "Threads"
	CHECKERS = "Checkerboard"
	DIAMONDS = "Diamonds"
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type Model struct {
	GameEngine *CGL
	FPS        time.Duration
	PresetList list.Model
	GameState  int
	EditState  int
	Height     int
	Width      int
}

type TickMsg struct{}

func frameTick(fps time.Duration) tea.Cmd {
	return tea.Tick(time.Second/fps, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m Model) Init() tea.Cmd {
	return tea.Sequence([]tea.Cmd{
		tea.SetWindowTitle("Conway's Game of Life"),
		frameTick(m.FPS),
	}...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.GameState == Playing {
				return m, tea.Quit
			} else if m.GameState == Mapping {
				return m, tea.Quit
			} else if m.GameState == PresetChoosing {
				m.GameState = Mapping
			}
		case tea.KeyEnter:
			if m.GameState == Playing {
				break
			} else if m.GameState == Mapping {
				m.GameState = Playing
				m.GameEngine.StartGame()
				cmds = append(cmds, tea.DisableMouse, tea.ClearScreen)
			} else if m.GameState == PresetChoosing {
				choice, ok := m.PresetList.SelectedItem().(item)
				if ok {
					switch choice {
					case RAND:
						m.GameEngine.RandomFill()
					case EDGES:
						m.GameEngine.EdgeFill()
					case PILLARS:
						m.GameEngine.PillarFill()
					case ROWS:
						m.GameEngine.RowFill()
					case DOTTED:
						m.GameEngine.DottedLines()
					case THREADS:
						m.GameEngine.Threads()
					case CHECKERS:
						m.GameEngine.Checkerboard()
					case DIAMONDS:
						m.GameEngine.Diamonds(5)
					}
				}
				m.GameState = Mapping
			}
		case tea.KeySpace:
			if m.GameState == Playing {
				m.GameState = Mapping
				cmds = append(cmds, tea.EnableMouseCellMotion)
			} else if m.GameState == Mapping {
				m.GameState = PresetChoosing
			} else if m.GameState == PresetChoosing {
				break
			}
		case tea.KeyBackspace:
			if m.GameState == Playing {
				m.GameEngine.ResetMap()
				m.GameState = Mapping
				cmds = append(cmds, tea.EnableMouseCellMotion)
			} else if m.GameState == Mapping {
				m.GameEngine.ResetMap()
			} else if m.GameState == PresetChoosing {
				m.GameState = Mapping
			}
		case tea.KeyRight:
			m.FPS += 1
		case tea.KeyLeft:
			m.FPS -= 1
		}
	case tea.MouseMsg:
		if m.GameState != Mapping {
			break
		}
		switch msg.Action {
		case tea.MouseActionPress:
			switch msg.Button {
			case tea.MouseButton(tea.MouseButtonLeft):
				m.EditState = Adding
			case tea.MouseButton(tea.MouseButtonRight):
				m.EditState = Removing
			}
		case tea.MouseActionMotion:
			switch msg.Button {
			case tea.MouseButton(tea.MouseButtonLeft):
				if m.EditState == Adding {
					m.GameEngine.UpdateAdd(msg.Y-9, msg.X)
				}
			case tea.MouseButton(tea.MouseButtonRight):
				if m.EditState == Removing {
					m.GameEngine.UpdateRemove(msg.Y-9, msg.X)
				}
			}
		case tea.MouseActionRelease:
			switch msg.Button {
			case tea.MouseButton(tea.MouseButtonLeft):
				m.EditState = Observing
			case tea.MouseButton(tea.MouseButtonRight):
				m.EditState = Observing
			}
		}
	case tea.WindowSizeMsg:
		m.Height = msg.Height - HEADING_SIZE
		m.Width = msg.Width
		m.PresetList.SetWidth(m.Width)
		m.GameEngine.Resize(m.Height, m.Width)
	case TickMsg:
		return m, frameTick(m.FPS)
	}
	if m.GameState == PresetChoosing {
		var cmd tea.Cmd
		m.PresetList, cmd = m.PresetList.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	frame := strings.Builder{}
	for h := 0; h < m.Height; h++ {
		for w := 0; w < m.Width; w++ {
			if m.GameEngine.GetCell(h, w) {
				frame.WriteString(colors[0].Render("■"))
			} else {
				frame.WriteString(" ")
			}
		}
		frame.WriteRune('\n')
	}
	var titleMsg string
	if m.GameState == Playing {
		//sync frame render to game state
		m.GameEngine.SyncFrame()
		titleMsg = TITLE
	} else if m.GameState == Mapping {
		titleMsg = `MAP EDITOR
LMB draw/RMB erase
SPACE: choose fill preset
BACKSPACE: reset
ENTER: draw life!
        `
	} else if m.GameState == PresetChoosing {
		titleMsg = fmt.Sprintf("MAP EDITOR\n%s", m.PresetList.View())
	}

	return fmt.Sprintf(
		`%s
%s
%s
%s
%s`,
		colors[1].Width(m.Width).AlignHorizontal(0.5).Render(titleMsg),
		colors[2].Width(m.Width).Render(strings.Repeat("=", m.Width)),
		colors[2].Width(m.Width).AlignHorizontal(0.5).Render(fmt.Sprintf("FPS: %d  ←-/+→", m.FPS)),
		colors[2].Width(m.Width).AlignHorizontal(0.5).Render("Press Esc/Ctrl+C to quit"),
		frame.String(),
	)
}

func InitModel(gameEngine *CGL, height int, width int) Model {
	m := Model{
		GameEngine: gameEngine,
		GameState:  Mapping,
		PresetList: list.New([]list.Item{
			item(RAND),
			item(EDGES),
			item(PILLARS),
			item(ROWS),
			item(DOTTED),
			item(THREADS),
			item(CHECKERS),
			item(DIAMONDS),
		},
			itemDelegate{},
			width, 5,
		),
		FPS:       10,
		EditState: Observing,
		Height:    height,
		Width:     width,
	}
	m.PresetList.SetShowHelp(false)
	m.PresetList.SetShowTitle(false)
	m.PresetList.SetShowStatusBar(false)
	m.PresetList.SetFilteringEnabled(false)
	m.PresetList.DisableQuitKeybindings()
	m.PresetList.Styles.PaginationStyle = paginationStyle
	return m
}
