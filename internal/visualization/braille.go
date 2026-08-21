package visualization

import (
	"os"
	"strings"
)

// SimpleOrbitsEnvVar disables braille orbit rendering in favour of a single '·' glyph
const SimpleOrbitsEnvVar = "GO_SOLAR_SYSTEM_SIMPLE_ORBITS"

const (
	brailleBase = 0x2800

	// BrailleDotsX and BrailleDotsY are the sub-cell resolution of one braille character
	BrailleDotsX = 2
	BrailleDotsY = 4
)

// brailleDotBits maps a sub-cell coordinate to its Unicode braille bit
var brailleDotBits = [BrailleDotsX][BrailleDotsY]byte{
	{0x01, 0x02, 0x04, 0x40},
	{0x08, 0x10, 0x20, 0x80},
}

// SimpleOrbitsRequested reports whether the environment asks for plain orbits
func SimpleOrbitsRequested() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(SimpleOrbitsEnvVar))) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// BrailleCanvas accumulates sub-cell dots for a character grid
type BrailleCanvas struct {
	width  int
	height int
	cells  []byte
}

// NewBrailleCanvas creates a canvas covering a width x height character grid
func NewBrailleCanvas(width, height int) *BrailleCanvas {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return &BrailleCanvas{
		width:  width,
		height: height,
		cells:  make([]byte, width*height),
	}
}

// Set marks a sub-cell dot, ignoring coordinates outside the canvas
func (bc *BrailleCanvas) Set(subX, subY int) {
	if subX < 0 || subY < 0 {
		return
	}

	cellX := subX / BrailleDotsX
	cellY := subY / BrailleDotsY
	if cellX >= bc.width || cellY >= bc.height {
		return
	}

	bc.cells[cellY*bc.width+cellX] |= brailleDotBits[subX%BrailleDotsX][subY%BrailleDotsY]
}

// CompositeInto writes the accumulated dots into grid, leaving occupied cells untouched
func (bc *BrailleCanvas) CompositeInto(grid [][]rune) {
	for y := 0; y < bc.height && y < len(grid); y++ {
		for x := 0; x < bc.width && x < len(grid[y]); x++ {
			mask := bc.cells[y*bc.width+x]
			if mask == 0 || grid[y][x] != ' ' {
				continue
			}

			grid[y][x] = rune(brailleBase | int(mask))
		}
	}
}

// IsBraille reports whether r is a braille pattern character
func IsBraille(r rune) bool {
	return r >= brailleBase && r <= brailleBase+0xFF
}
