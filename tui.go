package main

import (
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	TITLE = `
  _____                                _____                      ___  __   _ ___   
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
	Mapping = 0
	Playing = 1
    PresetChoosing = 2
	// Edit State
	Observing = 0
	Removing  = 1
	Adding    = 2
	// Preset choices
	RAND    = "Random Fill"
	EDGES   = "Edge tracing"
	PILLARS = "Pillars"
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
	GameEngine   *CGL
	FPS          time.Duration
	PresetList   list.Model
	GameState    int
	EditState    int
	Height       int
	Width        int
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
                return m, tea.Batch(tea.DisableMouse, tea.ClearScreen)
			} else if m.GameState == PresetChoosing {
                choice, ok := m.PresetList.SelectedItem().(item)
                if ok {
                    if choice == RAND {
                        m.GameEngine.RandomFill()
                    } else if choice == EDGES {
                        m.GameEngine.EdgeFill()
                    } else if choice == PILLARS {
                        m.GameEngine.PillarFill()
                    }
                }
                m.GameState = Mapping
            }
		case tea.KeySpace:
			if m.GameState == Playing {
                m.GameState = Mapping
                return m, tea.EnableMouseCellMotion
			} else if m.GameState == Mapping {
                m.GameState = PresetChoosing
            } else if m.GameState == PresetChoosing {
                break
            }
		case tea.KeyBackspace:
			if m.GameState == Playing {
                m.GameEngine.ResetMap()
                m.GameState = Mapping
                return m, tea.EnableMouseCellMotion
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
		m.Height = int(math.Min(float64(H_MAX), float64(msg.Height-10)))
		m.Width = int(math.Min(float64(W_MAX), float64(msg.Width)))
        m.PresetList.SetWidth(m.Width)
	case TickMsg:
		return m, frameTick(m.FPS)
	default:
	}
    if m.GameState == PresetChoosing {
        var cmd tea.Cmd
        m.PresetList, cmd = m.PresetList.Update(msg)
        return m, cmd
    }
	return m, nil
}

func (m Model) View() string {
	frame := strings.Builder{}
	gameMap := *m.GameEngine.GetReadOnlyMap()
	for h := 0; h < m.Height; h++ {
		for w := 0; w < m.Width; w++ {
			if gameMap[h][w] {
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
			item(PILLARS)},
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
