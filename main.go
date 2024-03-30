package main

import (
	"fmt"
	"math/rand"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	H_MAX = 66
	W_MAX = 160
)

type CGL struct {
	gameMap  [][]bool
	updateCh chan struct{}
	height   int
	width    int
}

func initCGL() CGL {
	cgl := CGL{
		gameMap:  make([][]bool, H_MAX),
		updateCh: make(chan struct{}),
		height:   H_MAX,
		width:    W_MAX,
	}
	for i := 0; i < cgl.height; i++ {
		cgl.gameMap[i] = make([]bool, cgl.width)
	}
	return cgl
}

func (cgl *CGL) neighbors(gameMap [][]bool, r int, c int) int {
	total := 0
	var adr, bdr, dc int
	if r > 0 {
		adr = r - 1
	} else {
		adr = cgl.height - 1
	}
	bdr = (r + 1) % cgl.height
	if c > 0 {
		dc = c - 1
	} else {
		dc = cgl.width - 1
	}
	if gameMap[r][dc] {
		total += 1
	}
	for i := 0; i < 3; i++ {
		if gameMap[adr][dc] {
			total += 1
		}
		if gameMap[bdr][dc] {
			total += 1
		}
		dc = (dc + 1) % cgl.width
	}
	if gameMap[r][(c+1)%cgl.width] {
		total += 1
	}
	return total
}

func (cgl *CGL) gameLoop() {
	for {
		curr_map := make([][]bool, cgl.height)
		for i := range cgl.gameMap {
			curr_map[i] = make([]bool, cgl.width)
			copy(curr_map[i], cgl.gameMap[i])
		}
		for r := 0; r < cgl.height; r++ {
			for c := 0; c < cgl.width; c++ {
				n := cgl.neighbors(curr_map, r, c)
				//Live cell
				if curr_map[r][c] {
					if n < 2 || n > 3 {
						cgl.gameMap[r][c] = false
					}
					//Dead cell
				} else {
					if n == 3 {
						cgl.gameMap[r][c] = true
					}
				}
			}
		}
		<-cgl.updateCh
	}
}

func (cgl *CGL) RandomFill() {
	for i := 0; i < cgl.height; i++ {
		for j := 0; j < cgl.width; j++ {
			v := rand.Intn(8) // 1/8 chance to alive
			if v == 0 {
				cgl.gameMap[i][j] = true
			} else {
				cgl.gameMap[i][j] = false
			}
		}
	}
}

func (cgl *CGL) EdgeFill() {
	for i := 0; i < cgl.height; i++ {
		for j := 0; j < cgl.width; j++ {
			if i == 0 || i == cgl.height-1 || j == 0 || j == cgl.width-1 {
				cgl.gameMap[i][j] = true
			}
		}
	}
}

func (cgl *CGL) PillarFill() {
	startP1 := (cgl.width / 3) - 1
	startP2 := startP1 * 2
	for i := 0; i < cgl.height; i++ {
		for j := startP1; j <= startP1+3; j++ {
			cgl.gameMap[i][j] = true
		}
		for j := startP2; j <= startP2+3; j++ {
			cgl.gameMap[i][j] = true
		}
	}
}

func (cgl *CGL) RowFill() {
	startP1 := (cgl.height / 3) - 1
	startP2 := startP1 * 2
	for j := 0; j < cgl.width; j++ {
		for i := startP1; i <= startP1+3; i++ {
			cgl.gameMap[i][j] = true
		}
		for i := startP2; i <= startP2+3; i++ {
			cgl.gameMap[i][j] = true
		}
	}
}

func (cgl *CGL) DottedLines() {
	for i := 0; i < cgl.height; i += 3 {
		for j := 0; j < cgl.width; j += 3 {
			if (i+j)%2 == 0 {
				cgl.UpdateAdd(i, j)
				cgl.UpdateAdd(i, j+1)
				cgl.UpdateAdd(i, j+2)
			} else {
				cgl.UpdateRemove(i, j)
				cgl.UpdateRemove(i, j+1)
				cgl.UpdateRemove(i, j+2)
			}
		}
	}
}

func (cgl *CGL) Threads() {
	for i := 0; i < cgl.height; i++ {
		for j := 0; j < cgl.width; j += 3 {
			if (i+j)%2 == 0 {
				cgl.UpdateAdd(i, j)
				cgl.UpdateAdd(i, j+1)
				cgl.UpdateAdd(i, j+2)
			} else {
				cgl.UpdateRemove(i, j)
				cgl.UpdateRemove(i, j+1)
				cgl.UpdateRemove(i, j+2)
			}
		}
	}
}

func (cgl *CGL) Checkerboard() {
	prev := true
	for i := 0; i < cgl.height; i++ {
		if i != 0 && i%4 == 0 {
			prev = !prev
		}
		for j := 0; j < cgl.width; j += 4 {
			if prev {
				cgl.UpdateAdd(i, j)
				cgl.UpdateAdd(i, j+1)
				cgl.UpdateAdd(i, j+2)
				cgl.UpdateAdd(i, j+3)
				prev = false
			} else {
				cgl.UpdateRemove(i, j)
				cgl.UpdateRemove(i, j+1)
				cgl.UpdateRemove(i, j+2)
				cgl.UpdateRemove(i, j+3)
				prev = true
			}
		}
	}
}

func (cgl *CGL) Diamonds(density int) {
	delta := cgl.height / density
	for h := 0; h <= cgl.height; h += delta {
		for j := 0; j < cgl.width; j++ {
			for i := 0; i < delta; i++ {
				cgl.UpdateAdd(h+i, j)
				cgl.UpdateAdd(h+i, j+1)
				cgl.UpdateAdd(h+delta-1-i, j)
				cgl.UpdateAdd(h+delta-1-i, j+1)
				j++
			}
		}
	}
}

func (cgl *CGL) ResetMap() {
	for i := 0; i < cgl.height; i++ {
		for j := 0; j < cgl.width; j++ {
			cgl.gameMap[i][j] = false
		}
	}
}

func (cgl *CGL) UpdateAdd(x, y int) {
	if x < 0 || x >= cgl.height {
		return
	}
	if y < 0 || y >= cgl.width {
		return
	}
	cgl.gameMap[x][y] = true
}

func (cgl *CGL) UpdateRemove(x, y int) {
	if x < 0 || x >= cgl.height {
		return
	}
	if y < 0 || y >= cgl.width {
		return
	}
	cgl.gameMap[x][y] = false
}

func (cgl *CGL) SyncFrame() {
	cgl.updateCh <- struct{}{}
}

func (cgl *CGL) GetReadOnlyMap() *[][]bool {
	return &cgl.gameMap
}

func (cgl *CGL) StartGame() {
	go cgl.gameLoop()
}

func main() {
	cgl := initCGL()
	tui_model := InitModel(&cgl, cgl.height, cgl.width)
	p := tea.NewProgram(
		tui_model,
		tea.WithMouseCellMotion(),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
