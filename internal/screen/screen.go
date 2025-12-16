package screen

import (
	"github.com/gdamore/tcell/v2"
)

// QBasic color palette (16 colors)
var QBasicColors = []tcell.Color{
	tcell.ColorBlack,        // 0
	tcell.ColorNavy,         // 1
	tcell.ColorGreen,        // 2
	tcell.ColorTeal,         // 3
	tcell.ColorMaroon,       // 4
	tcell.ColorPurple,       // 5
	tcell.ColorOlive,        // 6
	tcell.ColorSilver,       // 7
	tcell.ColorGray,         // 8
	tcell.ColorBlue,         // 9
	tcell.ColorLime,         // 10
	tcell.ColorAqua,         // 11
	tcell.ColorRed,          // 12
	tcell.ColorFuchsia,      // 13
	tcell.ColorYellow,       // 14
	tcell.ColorWhite,        // 15
}

// Screen represents the terminal display for BASIC programs
type Screen struct {
	tcell    tcell.Screen
	rows     int
	cols     int
	cursorX  int
	cursorY  int
	fgColor  int
	bgColor  int
	style    tcell.Style
	keyQueue chan string
}

// New creates a new Screen
func New() (*Screen, error) {
	tscreen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := tscreen.Init(); err != nil {
		return nil, err
	}

	cols, rows := tscreen.Size()

	s := &Screen{
		tcell:    tscreen,
		rows:     rows,
		cols:     cols,
		cursorX:  0,
		cursorY:  0,
		fgColor:  7, // white
		bgColor:  0, // black
		keyQueue: make(chan string, 16),
	}
	s.updateStyle()

	return s, nil
}

// Close shuts down the screen
func (s *Screen) Close() {
	if s.tcell != nil {
		s.tcell.Fini()
	}
}

func (s *Screen) updateStyle() {
	fg := tcell.ColorWhite
	bg := tcell.ColorBlack

	if s.fgColor >= 0 && s.fgColor < len(QBasicColors) {
		fg = QBasicColors[s.fgColor]
	}
	if s.bgColor >= 0 && s.bgColor < len(QBasicColors) {
		bg = QBasicColors[s.bgColor]
	}

	s.style = tcell.StyleDefault.Foreground(fg).Background(bg)
}

// Print prints a string at the current cursor position
func (s *Screen) Print(str string) {
	for _, ch := range str {
		if ch == '\n' {
			s.cursorX = 0
			s.cursorY++
			if s.cursorY >= s.rows {
				s.scroll()
				s.cursorY = s.rows - 1
			}
		} else if ch == '\r' {
			s.cursorX = 0
		} else if ch == '\a' {
			// Bell - do nothing in terminal
		} else {
			if s.cursorX < s.cols {
				s.tcell.SetContent(s.cursorX, s.cursorY, ch, nil, s.style)
				s.cursorX++
			}
			if s.cursorX >= s.cols {
				s.cursorX = 0
				s.cursorY++
				if s.cursorY >= s.rows {
					s.scroll()
					s.cursorY = s.rows - 1
				}
			}
		}
	}
	s.tcell.Show()
}

// Println prints a string followed by a newline
func (s *Screen) Println(str string) {
	s.Print(str + "\n")
}

// Clear clears the screen
func (s *Screen) Clear() {
	s.tcell.Clear()
	s.cursorX = 0
	s.cursorY = 0
	s.tcell.Show()
}

// Locate moves the cursor to the specified position (1-indexed)
func (s *Screen) Locate(row, col int) {
	s.cursorY = row - 1
	s.cursorX = col - 1

	if s.cursorY < 0 {
		s.cursorY = 0
	}
	if s.cursorY >= s.rows {
		s.cursorY = s.rows - 1
	}
	if s.cursorX < 0 {
		s.cursorX = 0
	}
	if s.cursorX >= s.cols {
		s.cursorX = s.cols - 1
	}
}

// SetColor sets foreground and background colors
func (s *Screen) SetColor(fg, bg int) {
	s.fgColor = fg
	s.bgColor = bg
	s.updateStyle()
}

// GetKey returns a keypress (non-blocking, returns "" if no key)
func (s *Screen) GetKey() string {
	select {
	case key := <-s.keyQueue:
		return key
	default:
		return ""
	}
}

// GetSize returns the screen dimensions
func (s *Screen) GetSize() (rows, cols int) {
	return s.rows, s.cols
}

// scroll scrolls the screen up one line
func (s *Screen) scroll() {
	// Move all lines up
	for y := 0; y < s.rows-1; y++ {
		for x := 0; x < s.cols; x++ {
			mainc, combc, style, _ := s.tcell.GetContent(x, y+1)
			s.tcell.SetContent(x, y, mainc, combc, style)
		}
	}
	// Clear bottom line
	for x := 0; x < s.cols; x++ {
		s.tcell.SetContent(x, s.rows-1, ' ', nil, s.style)
	}
}

// PollEvent polls for an event and queues keypresses
func (s *Screen) PollEvent() tcell.Event {
	ev := s.tcell.PollEvent()
	if keyEv, ok := ev.(*tcell.EventKey); ok {
		key := keyEventToString(keyEv)
		if key != "" {
			select {
			case s.keyQueue <- key:
			default:
				// Queue full, drop key
			}
		}
	}
	return ev
}

// Sync synchronizes the screen
func (s *Screen) Sync() {
	s.tcell.Sync()
}

// Show updates the display
func (s *Screen) Show() {
	s.tcell.Show()
}

// SetCell sets a cell directly
func (s *Screen) SetCell(x, y int, ch rune) {
	if x >= 0 && x < s.cols && y >= 0 && y < s.rows {
		s.tcell.SetContent(x, y, ch, nil, s.style)
	}
}

func keyEventToString(ev *tcell.EventKey) string {
	if ev.Key() == tcell.KeyRune {
		return string(ev.Rune())
	}

	switch ev.Key() {
	case tcell.KeyUp:
		return "\x00H" // Extended key code
	case tcell.KeyDown:
		return "\x00P"
	case tcell.KeyLeft:
		return "\x00K"
	case tcell.KeyRight:
		return "\x00M"
	case tcell.KeyEnter:
		return "\r"
	case tcell.KeyEscape:
		return "\x1B"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "\b"
	case tcell.KeyTab:
		return "\t"
	case tcell.KeyF1:
		return "\x00;"
	case tcell.KeyF2:
		return "\x00<"
	case tcell.KeyF3:
		return "\x00="
	case tcell.KeyF4:
		return "\x00>"
	case tcell.KeyF5:
		return "\x00?"
	case tcell.KeyF6:
		return "\x00@"
	case tcell.KeyF7:
		return "\x00A"
	case tcell.KeyF8:
		return "\x00B"
	case tcell.KeyF9:
		return "\x00C"
	case tcell.KeyF10:
		return "\x00D"
	case tcell.KeyHome:
		return "\x00G"
	case tcell.KeyEnd:
		return "\x00O"
	case tcell.KeyPgUp:
		return "\x00I"
	case tcell.KeyPgDn:
		return "\x00Q"
	case tcell.KeyInsert:
		return "\x00R"
	case tcell.KeyDelete:
		return "\x00S"
	default:
		return ""
	}
}

// DrawBox draws a box using Unicode box characters
func (s *Screen) DrawBox(x, y, width, height int, style tcell.Style) {
	// Corners
	s.tcell.SetContent(x, y, '┌', nil, style)
	s.tcell.SetContent(x+width-1, y, '┐', nil, style)
	s.tcell.SetContent(x, y+height-1, '└', nil, style)
	s.tcell.SetContent(x+width-1, y+height-1, '┘', nil, style)

	// Horizontal lines
	for i := x + 1; i < x+width-1; i++ {
		s.tcell.SetContent(i, y, '─', nil, style)
		s.tcell.SetContent(i, y+height-1, '─', nil, style)
	}

	// Vertical lines
	for i := y + 1; i < y+height-1; i++ {
		s.tcell.SetContent(x, i, '│', nil, style)
		s.tcell.SetContent(x+width-1, i, '│', nil, style)
	}
}

// FillRect fills a rectangle with a character
func (s *Screen) FillRect(x, y, width, height int, ch rune, style tcell.Style) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			s.tcell.SetContent(x+dx, y+dy, ch, nil, style)
		}
	}
}

// DrawText draws text at a position
func (s *Screen) DrawText(x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		if x+i < s.cols {
			s.tcell.SetContent(x+i, y, ch, nil, style)
		}
	}
}
