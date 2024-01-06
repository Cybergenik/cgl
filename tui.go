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
    colors = []lipgloss.Style {
        lipgloss.NewStyle().Foreground(lipgloss.Color("86")),
        lipgloss.NewStyle().Foreground(lipgloss.Color("201")),
        lipgloss.NewStyle().Foreground(lipgloss.Color("202")),
    }
)

type Model struct {
	Grid        *[][]bool
	UpdateCh    chan struct{}
    Height      int
    Width       int
}

type TickMsg struct{}

func frameTick() tea.Cmd {
	return func() tea.Msg { 
        <-time.After(time.Second/FPS) 
        return TickMsg{}
    }
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
        tea.ClearScreen,
        frameTick()}
	return tea.Sequence(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
        m.Height = int(math.Min(float64(H_MAX), float64(msg.Height)))
        m.Width = int(math.Min(float64(W_MAX), float64(msg.Width)))
	case TickMsg:
        return m, frameTick()
	}
	return m, nil
}

func (m Model) View() string {
    frame := strings.Builder{}
    for h:=0; h<m.Height; h++{
        for w:=0; w<m.Width; w++{
            if (*m.Grid)[h][w] {
                frame.WriteString(colors[0].
                    BorderTop(true).
                    BorderBottom(false).
                    BorderLeft(true).
                    BorderRight(false).
                    Render("■"))
            } else {
                frame.WriteString(colors[0].
                    BorderTop(true).
                    BorderBottom(false).
                    BorderLeft(true).
                    BorderRight(false).
                    Render(" "))
            }
        }
        frame.WriteRune('\n')
    }
    //sync frame render to game state
    m.UpdateCh<-struct{}{}
    return fmt.Sprintf(
        `
%s
%s
%s
        `,
        colors[1].Width(W_MAX).Render(TITLE),
        colors[2].Width(W_MAX).Render(strings.Repeat("=", W_MAX)),
        frame.String(),
    )
}

func InitModel(updateCh chan struct{}, gameMap *[][]bool, height int, width int) Model {
	m := Model{
        Grid: gameMap,
        UpdateCh: updateCh,
        Height: height,
        Width: width,
	}
	return m
}
