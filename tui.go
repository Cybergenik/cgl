package main

import (
	"fmt"
	"math"
	"strings"
	"time"

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
	cyan   = lipgloss.Color("86")
	purple = lipgloss.Color("201")
	orange = lipgloss.Color("202")
	colors = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(cyan),
		lipgloss.NewStyle().Foreground(purple),
		lipgloss.NewStyle().Foreground(orange),
	}
)

const (
	// Game State
	Mapping = 0
	Playing = 1
	// Edit State
	Observing = 0
	Removing  = 1
	Adding    = 2
)

type Model struct {
	GameEngine *CGL
	FPS        time.Duration
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
		tea.ClearScreen,
		frameTick(m.FPS),
	}...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.GameState != Mapping {
				break
			}
			m.GameState = Playing
			m.GameEngine.StartGame()
			return m, tea.Batch(tea.DisableMouse, tea.ClearScreen)
		case tea.KeySpace:
			if m.GameState != Mapping {
				break
			}
			m.GameEngine.RandomFill()
		case tea.KeyBackspace:
			if m.GameState != Mapping {
				break
			}
			m.GameEngine.ResetMap()
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
					m.GameEngine.UpdateAdd(msg.Y-10, msg.X)
				}
			case tea.MouseButton(tea.MouseButtonRight):
				if m.EditState == Removing {
					m.GameEngine.UpdateRemove(msg.Y-10, msg.X)
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
		m.Height = int(math.Min(float64(H_MAX), float64(msg.Height-12)))
		m.Width = int(math.Min(float64(W_MAX), float64(msg.Width)))
	case TickMsg:
		return m, frameTick(m.FPS)
	default:
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
		titleMsg = `MAP EDITOR: 
L-Click to Add, R-Click to Remove
SPACE:    generate a random map
BACKSPACE: reset
ENTER:    done
        `
	}

	return fmt.Sprintf(
		`
%s
%s
%s
%s
%s
`,
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
		FPS:        10,
		EditState:  Observing,
		Height:     height,
		Width:      width,
	}
	return m
}
