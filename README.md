<h1 align="center">Conway's Game of Life</h1>

<div align="center">
<h3>
in ASCII
</h3>

<img src="media/cgl_demo.gif" align="center" alt="visualization"/><br>

by: <a href="https://twitter.com/cybergenik" target="_blank">Luciano Remes</a>

<h3 align="center">Terminal based ASCII simulation</h3>

</div>


### Requirements:
- Golang version >= go1.25 linux/amd64

### Usage: 
1. `git clone https://github.com/Cybergenik/cgt && cd cgt`
2. `go run .`

##### Map Editor key bindings:
- <kbd>Left-MB</kbd>: draw
- <kbd>Right-MB</kbd>: erase
- <kbd>SPACE</kbd>: fill map with a preset (hjkl/←↓↑→, Enter, Backspace)
- <kbd>BACKSPACE</kbd>: clear map
- <kbd>ENTER</kbd>: draw life!

##### Simulation key bindings:
- <kbd>SPACE</kbd>: Pause the game state and go back to Map Editor
- <kbd>BACKSPACE</kbd>: Clear the map and go back to Map Editor

<kbd>Esc</kbd>/<kbd>Ctrl-C</kbd> to exit

_Run with DEFAULT=1 to set a default screen size of 160x66_
