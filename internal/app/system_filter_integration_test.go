package app

import (
	"github.com/mattn/go-runewidth"
	"io"
	"log"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/furan917/go-solar-system/internal/api"
	"github.com/furan917/go-solar-system/internal/systems"
	"github.com/furan917/go-solar-system/internal/visualization"
	"github.com/gdamore/tcell/v2"
)

func testSystemJSON(name, description string) string {
	return `{"systemName":"` + name + `","description":"` + description +
		`","bodies":[{"englishName":"star","bodyType":"Star"}]}`
}

func newTestDispatcher(t *testing.T) (*EventDispatcher, *AppState, *UIRenderer) {
	t.Helper()

	bundled := fstest.MapFS{
		"kepler-452.json":     {Data: []byte(testSystemJSON("Kepler-452", "Earth's cousin"))},
		"kepler-90.json":      {Data: []byte(testSystemJSON("Kepler-90", "Eight known planets"))},
		"trappist-1.json":     {Data: []byte(testSystemJSON("TRAPPIST-1", "Seven Earth-sized planets"))},
		"alpha-centauri.json": {Data: []byte(testSystemJSON("Alpha Centauri", "Closest binary system"))},
	}

	systemManager := systems.NewSystemManager(bundled)
	if err := systemManager.ScanSystems(); err != nil {
		t.Fatalf("ScanSystems: %v", err)
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("screen init: %v", err)
	}
	t.Cleanup(screen.Fini)
	screen.SetSize(140, 40)

	state := NewAppState()
	logger := log.New(io.Discard, "", 0)
	uiRenderer := NewUIRenderer(screen, visualization.NewRendererWithDefaults(140, 40), systemManager, state)
	errorHandler := NewErrorHandler(logger, state)
	planetService := NewPlanetService(api.NewClient(), systemManager)
	appSystemManager := NewSystemManager(state, planetService, uiRenderer, errorHandler, logger)
	mouseHandler := NewMouseEventHandler(state, uiRenderer, func() {}, func() {}, planetService, appSystemManager)

	return NewEventDispatcher(state, mouseHandler, appSystemManager, planetService, uiRenderer), state, uiRenderer
}

func press(ed *EventDispatcher, key tcell.Key, r rune) {
	ed.HandleEvent(tcell.NewEventKey(key, r, tcell.ModNone))
}

func typeText(ed *EventDispatcher, text string) {
	for _, r := range text {
		press(ed, tcell.KeyRune, r)
	}
}

func openSystemList(t *testing.T, ed *EventDispatcher, state *AppState) {
	t.Helper()

	press(ed, tcell.KeyRune, 's')
	if !state.ShowingSystemList {
		t.Fatal("system list did not open")
	}
}

func TestFilterNarrowsSystemListAsYouType(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	if len(ui.GetFilteredSystems()) != 5 {
		t.Fatalf("expected all systems before filtering, got %d", len(ui.GetFilteredSystems()))
	}

	press(ed, tcell.KeyRune, '/')
	if !state.SystemFilterActive {
		t.Fatal("'/' did not enter filter mode")
	}

	typeText(ed, "kepler")
	if state.SystemFilter != "kepler" {
		t.Fatalf("filter text = %q, want %q", state.SystemFilter, "kepler")
	}
	if got := len(ui.GetFilteredSystems()); got != 2 {
		t.Fatalf("expected 2 Kepler systems, got %d", got)
	}

	typeText(ed, "-90")
	if got := ui.GetFilteredSystems(); len(got) != 1 || got[0].Name != "kepler-90" {
		t.Fatalf("expected only kepler-90, got %v", got)
	}
}

func TestFilterModeCapturesCommandRunesAsText(t *testing.T) {
	ed, state, _ := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "qb")

	if state.SystemFilter != "qb" {
		t.Errorf("q/b should be filter text, got %q", state.SystemFilter)
	}
	if !state.IsRunning() {
		t.Error("'q' quit the application while filtering")
	}
	if !state.ShowingSystemList {
		t.Error("'b' closed the modal while filtering")
	}
}

func TestCtrlCQuitsWhileFiltering(t *testing.T) {
	ed, state, _ := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kep")
	press(ed, tcell.KeyCtrlC, 0)

	if state.IsRunning() {
		t.Error("Ctrl+C did not quit while filtering")
	}
}

func TestBackspaceWidensFilterAndKeepsSelectionValid(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kepler-90")
	if len(ui.GetFilteredSystems()) != 1 {
		t.Fatalf("setup: expected 1 match, got %d", len(ui.GetFilteredSystems()))
	}

	for i := 0; i < 3; i++ {
		press(ed, tcell.KeyBackspace2, 0)
	}

	if state.SystemFilter != "kepler" {
		t.Fatalf("filter = %q, want %q", state.SystemFilter, "kepler")
	}
	if got := len(ui.GetFilteredSystems()); got != 2 {
		t.Fatalf("expected list to widen back to 2, got %d", got)
	}
	assertSelectionInRange(t, state, len(ui.GetFilteredSystems()))
}

func TestNoMatchesLeavesSelectionSafeAndEnterDoesNothing(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "zzzz")

	if len(ui.GetFilteredSystems()) != 0 {
		t.Fatalf("expected no matches, got %d", len(ui.GetFilteredSystems()))
	}
	if state.SystemSelectedIndex != 0 || state.SystemScrollIndex != 0 {
		t.Errorf("indices not reset on empty list: %d/%d", state.SystemSelectedIndex, state.SystemScrollIndex)
	}

	press(ed, tcell.KeyEnter, 0)

	if !state.ShowingSystemList {
		t.Error("Enter on an empty filter closed the modal")
	}
	if state.CurrentSystem != "solar-system" {
		t.Errorf("Enter on an empty filter switched system to %q", state.CurrentSystem)
	}
}

func TestEscapeClearsFilterThenClosesModal(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kep")

	press(ed, tcell.KeyEscape, 0)
	if state.SystemFilter != "" || state.SystemFilterActive {
		t.Fatalf("first Escape did not clear the filter: %q active=%v", state.SystemFilter, state.SystemFilterActive)
	}
	if !state.ShowingSystemList {
		t.Fatal("first Escape closed the modal instead of clearing the filter")
	}
	if len(ui.GetFilteredSystems()) != 5 {
		t.Fatalf("clearing the filter did not restore the full list, got %d", len(ui.GetFilteredSystems()))
	}

	press(ed, tcell.KeyEscape, 0)
	if state.ShowingSystemList {
		t.Error("second Escape did not close the modal")
	}
}

func TestArrowKeysNavigateFilteredListOnly(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kepler")

	for i := 0; i < 5; i++ {
		press(ed, tcell.KeyDown, 0)
	}

	entries := ui.GetFilteredSystems()
	assertSelectionInRange(t, state, len(entries))
	if state.SystemSelectedIndex != len(entries)-1 {
		t.Errorf("selection %d should stop at the last match %d", state.SystemSelectedIndex, len(entries)-1)
	}
}

func TestFilterResetsWhenListReopens(t *testing.T) {
	ed, state, _ := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kep")
	press(ed, tcell.KeyRune, 'b')

	if !state.ShowingSystemList {
		t.Fatal("'b' closed the modal instead of being typed into the filter")
	}
	if state.SystemFilter != "kepb" {
		t.Fatalf("filter = %q, want %q", state.SystemFilter, "kepb")
	}

	press(ed, tcell.KeyEscape, 0)
	press(ed, tcell.KeyEscape, 0)
	openSystemList(t, ed, state)

	if state.SystemFilter != "" || state.SystemFilterActive {
		t.Errorf("filter survived reopening: %q active=%v", state.SystemFilter, state.SystemFilterActive)
	}
}

func assertSelectionInRange(t *testing.T, state *AppState, count int) {
	t.Helper()

	if count == 0 {
		if state.SystemSelectedIndex != 0 {
			t.Errorf("selection %d should be 0 for an empty list", state.SystemSelectedIndex)
		}
		return
	}

	if state.SystemSelectedIndex < 0 || state.SystemSelectedIndex >= count {
		t.Errorf("selection %d outside list of %d", state.SystemSelectedIndex, count)
	}
}

func clickAt(ed *EventDispatcher, x, y int) {
	ed.HandleEvent(tcell.NewEventMouse(x, y, tcell.Button1, tcell.ModNone))
}

func TestClickSelectsSystemInTheModalsLeftColumns(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	modalX, modalY, modalWidth, _ := ui.GetSystemModalDimensions(140, 40)
	narrowModalX, _, _, _ := ui.GetModalDimensions(140, 40)

	if modalX >= narrowModalX {
		t.Fatalf("expected the system modal to start left of the default modal: %d vs %d", modalX, narrowModalX)
	}
	if modalWidth != 80 {
		t.Fatalf("system modal width = %d, want 80", modalWidth)
	}

	entries := ui.GetFilteredSystems()
	targetRow := 2
	want := entries[targetRow].Name

	clickAt(ed, modalX+2, modalY+3+targetRow)

	if state.SystemSelectedIndex != targetRow {
		t.Errorf("click selected index %d, want %d", state.SystemSelectedIndex, targetRow)
	}
	if got := ui.GetSystemManager().GetCurrentSystem(); got != want {
		t.Errorf("click switched to %q, want %q", got, want)
	}
}

func TestClickMapsToFilteredEntryNotFullList(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kepler")

	entries := ui.GetFilteredSystems()
	if len(entries) != 2 {
		t.Fatalf("setup: expected 2 matches, got %d", len(entries))
	}

	modalX, modalY, _, _ := ui.GetSystemModalDimensions(140, 40)
	clickAt(ed, modalX+2, modalY+3+1)

	if got := ui.GetSystemManager().GetCurrentSystem(); got != entries[1].Name {
		t.Errorf("click switched to %q, want %q", got, entries[1].Name)
	}
}

func TestClickBelowTheLastMatchIsIgnored(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kepler")

	before := ui.GetSystemManager().GetCurrentSystem()
	modalX, modalY, _, _ := ui.GetSystemModalDimensions(140, 40)

	clickAt(ed, modalX+2, modalY+3+5)

	if got := ui.GetSystemManager().GetCurrentSystem(); got != before {
		t.Errorf("click past the last match switched to %q", got)
	}
	if !state.ShowingSystemList {
		t.Error("click past the last match closed the modal")
	}
}

func TestSystemModalRendersFilterLineAndMatches(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "kepler")

	ui.DrawScreen()

	modalX, modalY, modalWidth, _ := ui.GetSystemModalDimensions(140, 40)
	filterLine := readScreenLine(t, ui, modalX+2, modalY+2, modalWidth-4)

	if !strings.Contains(filterLine, "Filter: kepler") {
		t.Errorf("filter line = %q, want it to contain %q", filterLine, "Filter: kepler")
	}
	if !strings.Contains(filterLine, "2 matches") {
		t.Errorf("filter line = %q, want it to contain the match count", filterLine)
	}

	firstRow := readScreenLine(t, ui, modalX+2, modalY+3, modalWidth-4)
	if !strings.Contains(firstRow, "Kepler-452") {
		t.Errorf("first list row = %q, want it to show Kepler-452", firstRow)
	}
}

func TestSystemModalRendersEmptyFilterMessage(t *testing.T) {
	ed, state, ui := newTestDispatcher(t)
	openSystemList(t, ed, state)

	press(ed, tcell.KeyRune, '/')
	typeText(ed, "zzzz")

	ui.DrawScreen()

	modalX, modalY, modalWidth, _ := ui.GetSystemModalDimensions(140, 40)
	firstRow := readScreenLine(t, ui, modalX+2, modalY+3, modalWidth-4)

	if !strings.Contains(firstRow, "No systems match") {
		t.Errorf("first list row = %q, want the no-match message", firstRow)
	}
}

func readScreenLine(t *testing.T, ui *UIRenderer, x, y, width int) string {
	t.Helper()

	var line []rune
	for i := 0; i < width; i++ {
		primary, _, _, _ := ui.screen.GetContent(x+i, y)
		line = append(line, primary)
	}

	return string(line)
}

func TestStatusBarPositionsCurrentSystemByRuneCount(t *testing.T) {
	_, state, ui := newTestDispatcher(t)
	state.SetPlanets(nil)

	ui.DrawScreen()

	_, height := ui.screen.Size()
	line := readScreenLine(t, ui, 0, height-2, 140)

	instructions := "Arrow keys to navigate • Enter/Click to select • S for systems • Q to quit • 1-9 for direct selection"
	want := 2 + runewidth.StringWidth(instructions) + 3

	at := strings.Index(line, "• Current System:")
	if at < 0 {
		t.Fatalf("current system not found in status bar: %q", line)
	}

	got := runewidth.StringWidth(line[:at])
	if got != want {
		t.Errorf("current system starts at column %d, want %d", got, want)
	}

	if byteOffset := 2 + len(instructions) + 3; got == byteOffset {
		t.Errorf("column %d matches the byte length, so the offset is still byte-based", got)
	}
}

func TestStatusBarHasNoGapsAroundMultibyteSeparators(t *testing.T) {
	_, state, ui := newTestDispatcher(t)
	state.SetPlanets(nil)

	ui.DrawScreen()

	_, height := ui.screen.Size()
	line := readScreenLine(t, ui, 0, height-2, 140)

	if strings.Contains(line, "•   ") {
		t.Errorf("multibyte separator left blank columns behind it: %q", line)
	}
	if !strings.Contains(line, "navigate • Enter") {
		t.Errorf("expected single-space separators, got %q", line)
	}
}
