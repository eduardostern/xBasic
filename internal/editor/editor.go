package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/xbasic/xbasic/internal/interpreter"
	"github.com/xbasic/xbasic/internal/lexer"
	"github.com/xbasic/xbasic/internal/parser"
)

// QBasic-style colors
var (
	ColorMenuBg     = tcell.NewRGBColor(0, 0, 168)   // Dark blue
	ColorMenuFg     = tcell.ColorWhite
	ColorEditorBg   = tcell.NewRGBColor(0, 0, 168)   // Dark blue
	ColorEditorFg   = tcell.ColorWhite
	ColorStatusBg   = tcell.NewRGBColor(0, 168, 168) // Cyan
	ColorStatusFg   = tcell.ColorBlack
	ColorKeyword    = tcell.ColorWhite
	ColorString     = tcell.ColorYellow
	ColorComment    = tcell.ColorGray
	ColorNumber     = tcell.ColorAqua
	ColorLineNum    = tcell.ColorYellow
	ColorOutputBg   = tcell.ColorBlack
	ColorOutputFg   = tcell.ColorWhite
	ColorDialogBg   = tcell.ColorSilver
	ColorDialogFg   = tcell.ColorBlack
)

// Editor is the main IDE component
type Editor struct {
	app           *tview.Application
	mainFlex      *tview.Flex
	menuBar       *tview.TextView
	editorArea    *tview.TextArea
	statusBar     *tview.TextView
	outputArea    *tview.TextView
	showingOutput bool

	// Buffer state
	filename string
	modified bool
	lines    []string

	// Menu state
	menuActive    bool
	menuDropdown  *tview.List
	currentMenu   int
	menus         []Menu

	// Run state
	running     bool
	interpreter *interpreter.Interpreter
	outputText  strings.Builder
}

// Menu represents a dropdown menu
type Menu struct {
	Name  string
	Items []MenuItem
}

// MenuItem represents a menu item
type MenuItem struct {
	Name     string
	Shortcut string
	Action   func()
}

// New creates a new Editor
func New() *Editor {
	e := &Editor{
		app:      tview.NewApplication(),
		lines:    []string{""},
		filename: "Untitled",
	}

	e.setupMenus()
	e.setupUI()

	return e
}

func (e *Editor) setupMenus() {
	e.menus = []Menu{
		{
			Name: "File",
			Items: []MenuItem{
				{Name: "New", Shortcut: "Ctrl+N", Action: e.fileNew},
				{Name: "Open", Shortcut: "Ctrl+O", Action: e.fileOpen},
				{Name: "Save", Shortcut: "Ctrl+S", Action: e.fileSave},
				{Name: "Save As", Shortcut: "", Action: e.fileSaveAs},
				{Name: "─────────", Shortcut: "", Action: nil},
				{Name: "Exit", Shortcut: "Ctrl+Q", Action: e.fileExit},
			},
		},
		{
			Name: "Edit",
			Items: []MenuItem{
				{Name: "Cut", Shortcut: "Ctrl+X", Action: e.editCut},
				{Name: "Copy", Shortcut: "Ctrl+C", Action: e.editCopy},
				{Name: "Paste", Shortcut: "Ctrl+V", Action: e.editPaste},
				{Name: "─────────", Shortcut: "", Action: nil},
				{Name: "Select All", Shortcut: "Ctrl+A", Action: e.editSelectAll},
			},
		},
		{
			Name: "Run",
			Items: []MenuItem{
				{Name: "Start", Shortcut: "F5", Action: e.runProgram},
				{Name: "─────────", Shortcut: "", Action: nil},
				{Name: "Toggle Output", Shortcut: "F4", Action: e.toggleOutput},
			},
		},
		{
			Name: "Help",
			Items: []MenuItem{
				{Name: "About xBasic", Shortcut: "", Action: e.helpAbout},
			},
		},
	}
}

func (e *Editor) setupUI() {
	// Menu bar
	e.menuBar = tview.NewTextView()
	e.menuBar.SetDynamicColors(true)
	e.menuBar.SetBackgroundColor(ColorMenuBg)
	e.updateMenuBar()

	// Editor area
	e.editorArea = tview.NewTextArea()
	e.editorArea.SetBackgroundColor(ColorEditorBg)
	e.editorArea.SetTextStyle(tcell.StyleDefault.Foreground(ColorEditorFg).Background(ColorEditorBg))
	e.editorArea.SetPlaceholder("' Enter your BASIC code here...")
	e.editorArea.SetPlaceholderStyle(tcell.StyleDefault.Foreground(tcell.ColorGray).Background(ColorEditorBg))

	// Wrap editor in a frame for border
	editorFrame := tview.NewFrame(e.editorArea)
	editorFrame.SetBackgroundColor(ColorEditorBg)
	editorFrame.SetBorders(0, 0, 0, 0, 0, 0)

	// Status bar
	e.statusBar = tview.NewTextView()
	e.statusBar.SetDynamicColors(true)
	e.statusBar.SetBackgroundColor(ColorStatusBg)
	e.updateStatusBar()

	// Output area (hidden by default)
	e.outputArea = tview.NewTextView()
	e.outputArea.SetDynamicColors(true)
	e.outputArea.SetBackgroundColor(ColorOutputBg)
	e.outputArea.SetTextColor(ColorOutputFg)
	e.outputArea.SetScrollable(true)
	e.outputArea.SetTitle(" Output ")
	e.outputArea.SetBorder(true)

	// Main layout
	e.mainFlex = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.menuBar, 1, 0, false).
		AddItem(editorFrame, 0, 1, true).
		AddItem(e.statusBar, 1, 0, false)

	// Set up input capture
	e.app.SetInputCapture(e.handleInput)

	// Set changed handler for editor
	e.editorArea.SetChangedFunc(func() {
		e.modified = true
		e.updateStatusBar()
	})

	e.app.SetRoot(e.mainFlex, true)
}

func (e *Editor) updateMenuBar() {
	var menuStr strings.Builder
	menuStr.WriteString("[black:aqua] ")
	for i, menu := range e.menus {
		if i == e.currentMenu && e.menuActive {
			menuStr.WriteString(fmt.Sprintf("[black:white] %s [-:-] ", menu.Name))
		} else {
			menuStr.WriteString(fmt.Sprintf(" %s  ", menu.Name))
		}
	}
	menuStr.WriteString("[-:-]")
	e.menuBar.SetText(menuStr.String())
}

func (e *Editor) updateStatusBar() {
	row, col, _, _ := e.editorArea.GetCursor()
	modStr := ""
	if e.modified {
		modStr = " [*]"
	}
	status := fmt.Sprintf("[black:aqua] %s%s | Line %d, Col %d | F1=Help F5=Run F10=Menu [-:-]",
		filepath.Base(e.filename), modStr, row+1, col+1)
	e.statusBar.SetText(status)
}

func (e *Editor) handleInput(event *tcell.EventKey) *tcell.EventKey {
	// Handle menu navigation when menu is active
	if e.menuActive {
		return e.handleMenuInput(event)
	}

	// Global shortcuts
	switch event.Key() {
	case tcell.KeyF1:
		e.helpAbout()
		return nil
	case tcell.KeyF4:
		e.toggleOutput()
		return nil
	case tcell.KeyF5:
		e.runProgram()
		return nil
	case tcell.KeyF10:
		e.activateMenu()
		return nil
	case tcell.KeyCtrlN:
		e.fileNew()
		return nil
	case tcell.KeyCtrlO:
		e.fileOpen()
		return nil
	case tcell.KeyCtrlS:
		e.fileSave()
		return nil
	case tcell.KeyCtrlQ:
		e.fileExit()
		return nil
	case tcell.KeyEscape:
		if e.showingOutput {
			e.toggleOutput()
			return nil
		}
	}

	// Update status bar on cursor movement
	e.app.QueueUpdate(func() {
		e.updateStatusBar()
	})

	return event
}

func (e *Editor) handleMenuInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape, tcell.KeyF10:
		e.deactivateMenu()
		return nil
	case tcell.KeyLeft:
		e.currentMenu--
		if e.currentMenu < 0 {
			e.currentMenu = len(e.menus) - 1
		}
		e.updateMenuBar()
		e.showMenuDropdown()
		return nil
	case tcell.KeyRight:
		e.currentMenu++
		if e.currentMenu >= len(e.menus) {
			e.currentMenu = 0
		}
		e.updateMenuBar()
		e.showMenuDropdown()
		return nil
	case tcell.KeyEnter:
		if e.menuDropdown != nil {
			idx := e.menuDropdown.GetCurrentItem()
			if idx >= 0 && idx < len(e.menus[e.currentMenu].Items) {
				item := e.menus[e.currentMenu].Items[idx]
				e.deactivateMenu()
				if item.Action != nil {
					item.Action()
				}
			}
		}
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		// Let the list handle these
		return event
	}

	return event
}

func (e *Editor) activateMenu() {
	e.menuActive = true
	e.currentMenu = 0
	e.updateMenuBar()
	e.showMenuDropdown()
}

func (e *Editor) deactivateMenu() {
	e.menuActive = false
	e.updateMenuBar()
	if e.menuDropdown != nil {
		e.mainFlex.RemoveItem(e.menuDropdown)
		e.menuDropdown = nil
	}
	e.app.SetFocus(e.editorArea)
}

func (e *Editor) showMenuDropdown() {
	// Remove existing dropdown
	if e.menuDropdown != nil {
		e.mainFlex.RemoveItem(e.menuDropdown)
	}

	menu := e.menus[e.currentMenu]
	e.menuDropdown = tview.NewList()
	e.menuDropdown.SetBackgroundColor(ColorDialogBg)
	e.menuDropdown.SetMainTextColor(ColorDialogFg)
	e.menuDropdown.SetSelectedTextColor(ColorMenuFg)
	e.menuDropdown.SetSelectedBackgroundColor(ColorMenuBg)
	e.menuDropdown.ShowSecondaryText(true)
	e.menuDropdown.SetSecondaryTextColor(tcell.ColorGray)

	for _, item := range menu.Items {
		if item.Name == "─────────" {
			e.menuDropdown.AddItem("─────────", "", 0, nil)
		} else {
			e.menuDropdown.AddItem(item.Name, item.Shortcut, 0, nil)
		}
	}

	// Create modal with dropdown
	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 1, 0, false). // Space for menu bar
		AddItem(
			tview.NewFlex().
				AddItem(nil, e.currentMenu*10, 0, false). // Offset for menu position
				AddItem(e.menuDropdown, 20, 0, true).
				AddItem(nil, 0, 1, false),
			len(menu.Items)+2, 0, true,
		).
		AddItem(nil, 0, 1, false)

	pages := tview.NewPages().
		AddPage("main", e.mainFlex, true, true).
		AddPage("menu", modal, true, true)

	e.app.SetRoot(pages, true)
	e.app.SetFocus(e.menuDropdown)
}

// File operations

func (e *Editor) fileNew() {
	if e.modified {
		e.showConfirmDialog("Save changes?", func(yes bool) {
			if yes {
				e.fileSave()
			}
			e.doNewFile()
		})
	} else {
		e.doNewFile()
	}
}

func (e *Editor) doNewFile() {
	e.editorArea.SetText("", true)
	e.filename = "Untitled"
	e.modified = false
	e.updateStatusBar()
}

func (e *Editor) fileOpen() {
	e.showInputDialog("Open File", "Enter filename:", func(filename string) {
		if filename == "" {
			return
		}
		data, err := os.ReadFile(filename)
		if err != nil {
			e.showMessageDialog("Error", fmt.Sprintf("Cannot open file: %v", err))
			return
		}
		e.editorArea.SetText(string(data), true)
		e.filename = filename
		e.modified = false
		e.updateStatusBar()
	})
}

func (e *Editor) fileSave() {
	if e.filename == "Untitled" {
		e.fileSaveAs()
		return
	}
	e.doSave(e.filename)
}

func (e *Editor) fileSaveAs() {
	e.showInputDialog("Save As", "Enter filename:", func(filename string) {
		if filename == "" {
			return
		}
		// Add .bas extension if not present
		if !strings.HasSuffix(strings.ToLower(filename), ".bas") {
			filename += ".bas"
		}
		e.doSave(filename)
	})
}

func (e *Editor) doSave(filename string) {
	text := e.editorArea.GetText()
	err := os.WriteFile(filename, []byte(text), 0644)
	if err != nil {
		e.showMessageDialog("Error", fmt.Sprintf("Cannot save file: %v", err))
		return
	}
	e.filename = filename
	e.modified = false
	e.updateStatusBar()
}

func (e *Editor) fileExit() {
	if e.modified {
		e.showConfirmDialog("Save changes before exit?", func(yes bool) {
			if yes {
				e.fileSave()
			}
			e.app.Stop()
		})
	} else {
		e.app.Stop()
	}
}

// Edit operations

func (e *Editor) editCut() {
	// TextArea doesn't have built-in clipboard support
	// This would require implementing selection and clipboard
}

func (e *Editor) editCopy() {
	// Similar to editCut
}

func (e *Editor) editPaste() {
	// Similar to editCut
}

func (e *Editor) editSelectAll() {
	// Select all text
}

// Run operations

func (e *Editor) runProgram() {
	if e.running {
		return
	}

	// Get code from editor
	code := e.editorArea.GetText()
	if strings.TrimSpace(code) == "" {
		return
	}

	// Parse the code
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		e.showOutput()
		e.outputText.Reset()
		e.outputText.WriteString("=== Syntax Errors ===\n")
		for _, err := range p.Errors() {
			e.outputText.WriteString(err + "\n")
		}
		e.outputArea.SetText(e.outputText.String())
		return
	}

	// Create interpreter
	e.interpreter = interpreter.New(program)
	e.outputText.Reset()

	// Set up output
	e.interpreter.SetOutput(func(s string) {
		e.outputText.WriteString(s)
		e.app.QueueUpdateDraw(func() {
			e.outputArea.SetText(e.outputText.String())
			e.outputArea.ScrollToEnd()
		})
	})

	// Set up input
	e.interpreter.SetInput(func(prompt string) string {
		// For now, return empty string - proper input handling requires modal
		return ""
	})

	// Show output area
	e.showOutput()

	// Run in goroutine
	go func() {
		e.running = true
		e.outputText.WriteString("=== Running ===\n")
		e.app.QueueUpdateDraw(func() {
			e.outputArea.SetText(e.outputText.String())
		})

		err := e.interpreter.Run()

		e.running = false
		if err != nil {
			e.outputText.WriteString(fmt.Sprintf("\n=== Error ===\n%v\n", err))
		} else {
			e.outputText.WriteString("\n=== Done ===\n")
		}

		e.app.QueueUpdateDraw(func() {
			e.outputArea.SetText(e.outputText.String())
			e.outputArea.ScrollToEnd()
		})
	}()
}

func (e *Editor) toggleOutput() {
	if e.showingOutput {
		e.hideOutput()
	} else {
		e.showOutput()
	}
}

func (e *Editor) showOutput() {
	if e.showingOutput {
		return
	}
	e.showingOutput = true

	// Create split view
	e.mainFlex.Clear()
	e.mainFlex.AddItem(e.menuBar, 1, 0, false)

	split := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(e.editorArea, 0, 2, true).
		AddItem(e.outputArea, 0, 1, false)

	e.mainFlex.AddItem(split, 0, 1, true)
	e.mainFlex.AddItem(e.statusBar, 1, 0, false)
}

func (e *Editor) hideOutput() {
	if !e.showingOutput {
		return
	}
	e.showingOutput = false

	e.mainFlex.Clear()
	e.mainFlex.AddItem(e.menuBar, 1, 0, false)
	e.mainFlex.AddItem(e.editorArea, 0, 1, true)
	e.mainFlex.AddItem(e.statusBar, 1, 0, false)

	e.app.SetFocus(e.editorArea)
}

// Help

func (e *Editor) helpAbout() {
	e.showMessageDialog("About xBasic",
		"xBasic - QBasic Clone Interpreter\n\n"+
			"A cross-platform BASIC interpreter\n"+
			"for macOS and Linux.\n\n"+
			"Shortcuts:\n"+
			"F5  - Run program\n"+
			"F4  - Toggle output\n"+
			"F10 - Menu\n"+
			"Ctrl+S - Save\n"+
			"Ctrl+Q - Exit")
}

// Dialog helpers

func (e *Editor) showMessageDialog(title, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			e.app.SetRoot(e.mainFlex, true)
			e.app.SetFocus(e.editorArea)
		})
	modal.SetTitle(title)
	modal.SetBackgroundColor(ColorDialogBg)
	modal.SetTextColor(ColorDialogFg)

	e.app.SetRoot(modal, true)
}

func (e *Editor) showConfirmDialog(message string, callback func(bool)) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			e.app.SetRoot(e.mainFlex, true)
			e.app.SetFocus(e.editorArea)
			callback(buttonIndex == 0)
		})
	modal.SetBackgroundColor(ColorDialogBg)
	modal.SetTextColor(ColorDialogFg)

	e.app.SetRoot(modal, true)
}

func (e *Editor) showInputDialog(title, prompt string, callback func(string)) {
	input := tview.NewInputField().
		SetLabel(prompt + " ").
		SetFieldWidth(40).
		SetFieldBackgroundColor(tcell.ColorWhite).
		SetFieldTextColor(tcell.ColorBlack)

	form := tview.NewForm().
		AddFormItem(input).
		AddButton("OK", func() {
			e.app.SetRoot(e.mainFlex, true)
			e.app.SetFocus(e.editorArea)
			callback(input.GetText())
		}).
		AddButton("Cancel", func() {
			e.app.SetRoot(e.mainFlex, true)
			e.app.SetFocus(e.editorArea)
		})

	form.SetTitle(title)
	form.SetBorder(true)
	form.SetBackgroundColor(ColorDialogBg)
	form.SetButtonBackgroundColor(ColorMenuBg)
	form.SetButtonTextColor(ColorMenuFg)

	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)

	e.app.SetRoot(flex, true)
	e.app.SetFocus(input)
}

// Run starts the editor
func (e *Editor) Run() error {
	return e.app.Run()
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	e.editorArea.SetText(string(data), true)
	e.filename = filename
	e.modified = false
	e.updateStatusBar()
	return nil
}
