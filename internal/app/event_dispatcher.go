package app

import (
	"strconv"

	"github.com/furan917/go-solar-system/internal/models"
	"github.com/gdamore/tcell/v2"
)

type EventDispatcher struct {
	state         *AppState
	mouseHandler  *MouseEventHandler
	systemManager *SystemManager
	planetService *PlanetService
	uiRenderer    *UIRenderer
}

func NewEventDispatcher(state *AppState, mouseHandler *MouseEventHandler, systemManager *SystemManager, planetService *PlanetService, uiRenderer *UIRenderer) *EventDispatcher {
	return &EventDispatcher{
		state:         state,
		mouseHandler:  mouseHandler,
		systemManager: systemManager,
		planetService: planetService,
		uiRenderer:    uiRenderer,
	}
}

func (ed *EventDispatcher) HandleEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		ed.mouseHandler.HandleClick(ev)
	case *tcell.EventKey:
		ed.handleKeyboardEvent(ev)
	case *tcell.EventResize:
		ed.handleResizeEvent(ev)
	}
}

func (ed *EventDispatcher) handleKeyboardEvent(ev *tcell.EventKey) {
	if ed.state.IsShowingMoonDetails() {
		ed.handleMoonDetailsKeys(ev)
	} else if ed.state.IsShowingMoons() {
		ed.handleMoonListKeys(ev)
	} else if ed.state.IsShowingSystemList() {
		ed.handleSystemListKeys(ev)
	} else if ed.state.IsShowingDetails() {
		ed.handlePlanetDetailsKeys(ev)
	} else {
		ed.handleMainNavigationKeys(ev)
	}
}

func (ed *EventDispatcher) handleResizeEvent(ev *tcell.EventResize) {
	// Update renderer dimensions to handle scaling
	width, height := ev.Size()
	ed.uiRenderer.UpdateDimensions(width, height)
}

func (ed *EventDispatcher) handleMoonDetailsKeys(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyEnter:
		ed.state.ShowMoonList()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q', 'Q':
			ed.state.SetRunning(false)
		case 'b', 'B':
			ed.state.ShowMoonList()
		}
	default:
		// do nothing
	}
}

func (ed *EventDispatcher) handleMoonListKeys(ev *tcell.EventKey) {
	ed.handleMoonNavigation(ev)
}

func (ed *EventDispatcher) handleSystemListKeys(ev *tcell.EventKey) {
	ed.handleSystemNavigation(ev)
}

func (ed *EventDispatcher) handlePlanetDetailsKeys(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyEnter:
		ed.state.ResetModals()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q', 'Q', 'b', 'B':
			ed.state.ResetModals()
		case 'm', 'M':
			if len(ed.state.SelectedPlanet.Moons) > 0 {
				ed.state.ShowMoonList()
			}
		}
	default:
		// do nothing
	}
}

func (ed *EventDispatcher) handleMainNavigationKeys(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		ed.state.SetRunning(false)
	case tcell.KeyUp, tcell.KeyLeft:
		ed.navigatePlanet(-1)
	case tcell.KeyDown, tcell.KeyRight:
		ed.navigatePlanet(1)
	case tcell.KeyEnter:
		if ed.state.SelectedIndex < len(ed.state.GetPlanets()) {
			ed.showPlanetDetails(ed.state.GetPlanets()[ed.state.SelectedIndex])
		}
	case tcell.KeyRune:
		ed.handleMainNavigationRunes(ev.Rune())
	default:
		// do nothing
	}
}

func (ed *EventDispatcher) handleMainNavigationRunes(r rune) {
	switch r {
	case 'q', 'Q':
		ed.state.SetRunning(false)
	case 'h', 'H':
		// Help functionality placeholder
	case 's', 'S':
		ed.showSystemList()
	default:
		ed.handleDirectPlanetSelection(r)
	}
}

func (ed *EventDispatcher) navigatePlanet(direction int) {
	newIndex := ed.state.SelectedIndex + direction
	if newIndex >= 0 && newIndex < len(ed.state.GetPlanets()) {
		ed.state.UpdatePlanetSelection(newIndex, ed.state.GetPlanets()[newIndex])
	}
}

func (ed *EventDispatcher) handleDirectPlanetSelection(r rune) {
	if num, err := strconv.Atoi(string(r)); err == nil && num >= 1 && num <= len(ed.state.GetPlanets()) {
		newIndex := num - 1
		ed.state.UpdatePlanetSelection(newIndex, ed.state.GetPlanets()[newIndex])
		ed.showPlanetDetails(ed.state.GetPlanets()[newIndex])
	}
}

func (ed *EventDispatcher) showPlanetDetails(planet models.CelestialBody) {
	ed.state.ShowPlanetDetails(planet, ed.state.SelectedIndex)
}

func (ed *EventDispatcher) showSystemList() {
	ed.state.ShowingSystemList = true
	ed.state.ClearSystemFilter()
	ed.state.SystemScrollIndex = 0
	ed.state.SystemSelectedIndex = 0

	entries := ed.uiRenderer.GetFilteredSystems()
	currentSystem := ed.uiRenderer.GetSystemManager().GetCurrentSystem()
	for i, entry := range entries {
		if entry.Name == currentSystem {
			ed.state.SystemSelectedIndex = i
			break
		}
	}

	ed.state.ClampSystemSelection(len(entries))
}

func (ed *EventDispatcher) handleMoonNavigation(ev *tcell.EventKey) {
	moonCount := len(ed.state.SelectedPlanet.Moons)
	if moonCount == 0 {
		return
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		ed.state.ShowingMoons = false
		ed.state.ShowingDetails = true
	case tcell.KeyUp:
		if ed.state.MoonSelectedIndex > 0 {
			ed.state.MoonSelectedIndex--
			if ed.state.MoonSelectedIndex < ed.state.MoonScrollIndex {
				ed.state.MoonScrollIndex = ed.state.MoonSelectedIndex
			}
		}
	case tcell.KeyDown:
		if ed.state.MoonSelectedIndex < moonCount-1 {
			ed.state.MoonSelectedIndex++
			if ed.state.MoonSelectedIndex >= ed.state.MoonScrollIndex+10 {
				ed.state.MoonScrollIndex = ed.state.MoonSelectedIndex - 9
			}
		}
	case tcell.KeyEnter:
		ed.showMoonDetails()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q', 'Q':
			ed.state.SetRunning(false)
		case 'b', 'B':
			ed.state.ShowingMoons = false
			ed.state.ShowingDetails = true
		}
	default:
		// do nothing
	}
}

func (ed *EventDispatcher) handleSystemNavigation(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyCtrlC {
		ed.state.SetRunning(false)
		return
	}

	if ed.state.SystemFilterActive {
		ed.handleSystemFilterKeys(ev)
		return
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		ed.closeSystemList()
	case tcell.KeyUp:
		ed.navigateSystems(-1)
	case tcell.KeyDown:
		ed.navigateSystems(1)
	case tcell.KeyEnter:
		ed.systemManager.SwitchToSelectedSystem()
	case tcell.KeyRune:
		switch ev.Rune() {
		case '/':
			ed.state.SystemFilterActive = true
		case 'q', 'Q':
			ed.state.SetRunning(false)
		case 'b', 'B':
			ed.closeSystemList()
		}
	default:
		// do nothing
	}
}

// handleSystemFilterKeys consumes keys while the system filter is being typed
func (ed *EventDispatcher) handleSystemFilterKeys(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		ed.state.ClearSystemFilter()
		ed.clampSystemSelection()
	case tcell.KeyUp:
		ed.navigateSystems(-1)
	case tcell.KeyDown:
		ed.navigateSystems(1)
	case tcell.KeyEnter:
		ed.systemManager.SwitchToSelectedSystem()
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		ed.setSystemFilter(trimLastRune(ed.state.SystemFilter))
	case tcell.KeyRune:
		ed.setSystemFilter(ed.state.SystemFilter + string(ev.Rune()))
	default:
		// do nothing
	}
}

func (ed *EventDispatcher) setSystemFilter(filter string) {
	ed.state.SystemFilter = filter
	ed.clampSystemSelection()
}

func (ed *EventDispatcher) navigateSystems(direction int) {
	ed.state.HandleSystemNavigation(direction, len(ed.uiRenderer.GetFilteredSystems()))
}

func (ed *EventDispatcher) clampSystemSelection() {
	ed.state.ClampSystemSelection(len(ed.uiRenderer.GetFilteredSystems()))
}

func (ed *EventDispatcher) closeSystemList() {
	ed.state.ShowingSystemList = false
	ed.state.ClearSystemFilter()
}

func trimLastRune(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
}

func (ed *EventDispatcher) showMoonDetails() {
	if ed.state.MoonSelectedIndex < len(ed.state.SelectedPlanet.Moons) {
		moonData := ed.state.SelectedPlanet.Moons[ed.state.MoonSelectedIndex]
		moonHandler := ed.uiRenderer.GetRenderer().GetMoonHandler()
		moonName := moonHandler.GetMoonNameFromAPI(moonData)

		if moonData.ID != "" {
			if moonDetail, err := ed.planetService.GetClient().GetMoonData(moonData.ID); err == nil {
				ed.state.SelectedMoon = *moonDetail
				ed.state.SelectedMoon.BodyType = "Moon"
				ed.state.SelectedMoon.AroundPlanet = &models.Planet{
					EnglishName: ed.state.SelectedPlanet.EnglishName,
				}
			} else {
				ed.state.SelectedMoon = models.CelestialBody{
					ID:          moonData.ID,
					Name:        moonData.Name,
					EnglishName: moonName,
					BodyType:    "Moon",
					AroundPlanet: &models.Planet{
						EnglishName: ed.state.SelectedPlanet.EnglishName,
					},
				}
			}
		} else {
			ed.state.SelectedMoon = models.CelestialBody{
				ID:          moonData.ID,
				Name:        moonData.Name,
				EnglishName: moonName,
				BodyType:    "Moon",
				AroundPlanet: &models.Planet{
					EnglishName: ed.state.SelectedPlanet.EnglishName,
				},
			}
		}

		ed.state.ShowingMoonDetails = true
		ed.state.ShowingMoons = false
	}
}
