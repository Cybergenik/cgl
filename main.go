package main

import (
    "os"
    "fmt"
    "math/rand"

    tea "github.com/charmbracelet/bubbletea"
)

const (
    H_MAX = 65
    W_MAX = 125
)

type CGL struct {
    gameMap  [][]bool
    updateCh chan struct{}
    height   int
    width    int
}

func initCGL() CGL {
    cgl := CGL{
        gameMap: make([][]bool, H_MAX),
        updateCh: make(chan struct{}),
        height: H_MAX,
        width: W_MAX,
    }
    for i:=0; i < cgl.height; i++{
        cgl.gameMap[i] = make([]bool, cgl.width)
        for j:=0; j < cgl.width; j++{
            v := rand.Intn(10)
            if v == 0 {
                cgl.gameMap[i][j] = true
            } else {
                cgl.gameMap[i][j] = false
            }
        }
    }
    return cgl
}

func (cgl *CGL) neighbors(gameMap [][]bool, r int, c int) int{
    total := 0
    var adr, bdr, dc int
    if r > 0{
        adr = r-1
    } else {
        adr = cgl.height-1
    }
    bdr = (r+1)%cgl.height
    if c > 0{
        dc = c-1
    } else {
        dc = cgl.width-1
    }
    if gameMap[r][dc] {
        total += 1
    }
    for i:=0; i<3; i++ {
        if gameMap[adr][dc] {
            total += 1
        }
        if gameMap[bdr][dc] {
            total += 1
        }
        dc = (dc+1)%cgl.width
    }
    if gameMap[r][(c+1)%cgl.width] {
        total += 1
    }
    return total
}

func (cgl *CGL) gameLoop(){
    for {
        curr_map := make([][]bool, cgl.height)
        for i := range cgl.gameMap {
            curr_map[i] = make([]bool, cgl.width)
            copy(curr_map[i], cgl.gameMap[i])
        }
        for r:=0; r < cgl.height; r++{
            for c:=0; c < cgl.width; c++{
                n := cgl.neighbors(curr_map, r, c)
                //Live cell
                if curr_map[r][c] {
                    if n < 2 || n > 3{
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

func main() {
    cgl := initCGL()
    tui_model := InitModel(cgl.updateCh, &cgl.gameMap, cgl.height, cgl.width)
    go cgl.gameLoop()
    p := tea.NewProgram(tui_model)
    if _, err := p.Run(); err != nil {
        fmt.Println(err)
        os.Exit(-1)
    }
}
