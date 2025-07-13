package ui

import (
	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/gdamore/tcell/v2"
)

type Modal struct {
	screen       tcell.Screen
	x, y         int
	width        int
	height       int
	title        string
	content      []string
	borderStyle  tcell.Style
	titleStyle   tcell.Style
	contentStyle tcell.Style
}

type ModalConfig struct {
	Width    int
	Height   int
	Title    string
	Content  []string
	Position constants.ModalPosition
}

func NewModal(screen tcell.Screen, config ModalConfig) *Modal {
	screenWidth, screenHeight := screen.Size()

	var x, y int
	switch config.Position {
	case constants.TopRight:
		x = screenWidth - config.Width - 2
		y = 1
	case constants.Center:
		x = (screenWidth - config.Width) / 2
		y = (screenHeight - config.Height) / 2
	case constants.TopLeft:
		x = 2
		y = 1
	case constants.BottomRight:
		x = screenWidth - config.Width - 2
		y = screenHeight - config.Height - 2
	}

	return &Modal{
		screen:       screen,
		x:            x,
		y:            y,
		width:        config.Width,
		height:       config.Height,
		title:        config.Title,
		content:      config.Content,
		borderStyle:  tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue).Bold(true),
		titleStyle:   tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorDarkBlue).Bold(true),
		contentStyle: tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDarkBlue),
	}
}

func (m *Modal) Render() {
	m.drawBackground()
	m.drawBorder()
	m.drawTitle()
	m.drawContent()
}

func (m *Modal) drawBackground() {
	for y := m.y; y < m.y+m.height; y++ {
		for x := m.x; x < m.x+m.width; x++ {
			m.screen.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(tcell.ColorDarkBlue))
		}
	}
}

func (m *Modal) drawBorder() {
	// Top and bottom borders
	for x := m.x; x < m.x+m.width; x++ {
		m.screen.SetContent(x, m.y, '═', nil, m.borderStyle)
		m.screen.SetContent(x, m.y+m.height-1, '═', nil, m.borderStyle)
	}

	// Left and right borders
	for y := m.y; y < m.y+m.height; y++ {
		m.screen.SetContent(m.x, y, '║', nil, m.borderStyle)
		m.screen.SetContent(m.x+m.width-1, y, '║', nil, m.borderStyle)
	}

	// Corners
	m.screen.SetContent(m.x, m.y, '╔', nil, m.borderStyle)
	m.screen.SetContent(m.x+m.width-1, m.y, '╗', nil, m.borderStyle)
	m.screen.SetContent(m.x, m.y+m.height-1, '╚', nil, m.borderStyle)
	m.screen.SetContent(m.x+m.width-1, m.y+m.height-1, '╝', nil, m.borderStyle)
}

func (m *Modal) drawTitle() {
	if m.title != "" {
		m.drawTextAt(m.x+2, m.y+1, m.titleStyle, m.title)
	}
}

func (m *Modal) drawContent() {
	currentY := m.y + 3
	for _, line := range m.content {
		if currentY >= m.y+m.height-2 {
			break
		}
		m.drawTextAt(m.x+2, currentY, m.contentStyle, line)
		currentY++
	}
}

func (m *Modal) drawTextAt(x, y int, style tcell.Style, text string) {
	for i, r := range text {
		if x+i >= m.x+m.width-2 {
			break
		}
		m.screen.SetContent(x+i, y, r, nil, style)
	}
}

func (m *Modal) DrawInstructions(instructions string) {
	instructionStyle := tcell.StyleDefault.Foreground(tcell.ColorLightBlue).Background(tcell.ColorDarkBlue)
	m.drawTextAt(m.x+2, m.y+m.height-2, instructionStyle, instructions)
}
