package app

import (
    "math"
    "strings"

    "github.com/furan917/go-solar-system/internal/models"
    "github.com/gdamore/tcell/v2"
)

type MouseEventHandler struct {
    state           *AppState
    renderer        *UIRenderer
    showMoonList    func()
    showMoonDetails func()
    planetService   *PlanetService
    systemManager   *SystemManager
}

func NewMouseEventHandler(state *AppState, renderer *UIRenderer, showMoonList, showMoonDetails func(), planetService *PlanetService, systemManager *SystemManager) *MouseEventHandler {
    return &MouseEventHandler{
        state:           state,
        renderer:        renderer,
        showMoonList:    showMoonList,
        showMoonDetails: showMoonDetails,
        planetService:   planetService,
        systemManager:   systemManager,
    }
}

func (meh *MouseEventHandler) HandleClick(ev *tcell.EventMouse) {
    if ev.Buttons() != tcell.Button1 {
        return
    }

    mouseX, mouseY := ev.Position()

    if meh.handleInstructionBarClick(mouseX, mouseY) {
        return
    }

    if meh.handlePlanetListClick(mouseX, mouseY) {
        return
    }

    switch {
    case meh.state.ShowingMoonDetails:
        if meh.handleMoonDetailsModalClick(mouseX, mouseY) {
            return
        }
    case meh.state.ShowingMoons:
        if meh.handleMoonListModalClick(mouseX, mouseY) {
            return
        }
    case meh.state.ShowingSystemList:
        if meh.handleSystemListModalClick(mouseX, mouseY) {
            return
        }
    case meh.state.ShowingDetails:
        if meh.handlePlanetDetailsModalClick(mouseX, mouseY) {
            return
        }
    default:
        //
    }

    if meh.renderer.IsClickInModalArea(mouseX, mouseY) {
        return
    }

    for name, pos := range meh.state.GetPlanetPositions() {
        dx := float64(mouseX - pos.X)
        dy := float64(mouseY - pos.Y)
        distance := math.Sqrt(dx*dx + dy*dy)

        clickRadius := float64(pos.Radius + 2)
        if distance <= clickRadius {
            meh.state.SelectedPlanet = pos.Planet

            for i, planet := range meh.state.GetPlanets() {
                if planet.EnglishName == name {
                    meh.state.SelectedIndex = i
                    break
                }
            }

            if !meh.state.ShowingDetails && !meh.state.ShowingMoons && !meh.state.ShowingMoonDetails && !meh.state.ShowingSystemList {
                meh.state.ShowingDetails = true
            } else if meh.state.ShowingDetails {
            }
            return
        }
    }
}

func (meh *MouseEventHandler) handleInstructionBarClick(mouseX, mouseY int) bool {
    _, screenHeight := meh.renderer.screen.Size()
    instructionY := screenHeight - 2

    if mouseY != instructionY {
        return false
    }

    instructions := "Arrow keys to navigate • Enter/Click to select • S for systems • Q to quit • 1-9 for direct selection"

    sPos := strings.Index(instructions, "S for systems")
    if sPos >= 0 && mouseX >= 2+sPos && mouseX <= 2+sPos+13 {
        meh.state.ShowingSystemList = true
        meh.state.ShowingDetails = false
        meh.state.ShowingMoons = false
        meh.state.ShowingMoonDetails = false
        return true
    }

    qPos := strings.Index(instructions, "Q to quit")
    if qPos >= 0 && mouseX >= 2+qPos && mouseX <= 2+qPos+8 {
        meh.state.SetRunning(false)
        return true
    }

    return false
}

func (meh *MouseEventHandler) handleMoonDetailsModalClick(mouseX, mouseY int) bool {
    screenWidth, screenHeight := meh.renderer.screen.Size()
    contentLines := meh.renderer.calculateMoonDetailsLines(meh.state.SelectedMoon)
    dynamicHeight := minimum(contentLines+6, screenHeight-4)
    modalX, modalY, modalWidth, modalHeight := meh.renderer.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)

    if mouseX < modalX || mouseX >= modalX+modalWidth || mouseY < modalY || mouseY >= modalY+modalHeight {
        return false
    }

    instructionY := modalY + modalHeight - 2
    if mouseY == instructionY {
        meh.state.ShowingMoonDetails = false
        meh.state.ShowingMoons = true
        return true
    }

    return true
}

func (meh *MouseEventHandler) handleMoonListModalClick(mouseX, mouseY int) bool {
    screenWidth, screenHeight := meh.renderer.screen.Size()
    modalX, modalY, modalWidth, modalHeight := meh.renderer.GetModalDimensions(screenWidth, screenHeight)

    if mouseX < modalX || mouseX >= modalX+modalWidth || mouseY < modalY || mouseY >= modalY+modalHeight {
        return false
    }

    moonListStartY := modalY + 3
    maxVisibleMoons := 10

    if mouseY >= moonListStartY && mouseY < moonListStartY+maxVisibleMoons {
        moonIndex := meh.state.MoonScrollIndex + (mouseY - moonListStartY)
        if moonIndex < len(meh.state.SelectedPlanet.Moons) {
            meh.state.MoonSelectedIndex = moonIndex
            meh.showMoonDetailsInternal()
            return true
        }
    }

    instructionY := modalY + modalHeight - 2
    if mouseY == instructionY {
        meh.state.ShowingMoons = false
        meh.state.ShowingDetails = true
        return true
    }

    return true
}

func (meh *MouseEventHandler) handleSystemListModalClick(mouseX, mouseY int) bool {
    screenWidth, screenHeight := meh.renderer.screen.Size()
    modalX, modalY, modalWidth, modalHeight := meh.renderer.GetModalDimensions(screenWidth, screenHeight)

    if mouseX < modalX || mouseX >= modalX+modalWidth || mouseY < modalY || mouseY >= modalY+modalHeight {
        return false
    }

    systemListStartY := modalY + 3
    maxVisibleSystems := 12

    if mouseY >= systemListStartY && mouseY < systemListStartY+maxVisibleSystems {
        systemIndex := meh.state.SystemScrollIndex + (mouseY - systemListStartY)
        availableSystems := meh.renderer.GetSystemManager().GetAvailableSystems()

        if systemIndex < len(availableSystems) {
            meh.state.SystemSelectedIndex = systemIndex
            meh.systemManager.SwitchToSelectedSystem()
            return true
        }
    }

    instructionY := modalY + modalHeight - 2
    if mouseY == instructionY {
        meh.state.ShowingSystemList = false
        return true
    }

    return true
}

func (meh *MouseEventHandler) handlePlanetDetailsModalClick(mouseX, mouseY int) bool {
    screenWidth, screenHeight := meh.renderer.screen.Size()
    contentLines := meh.renderer.calculatePlanetDetailsLines(meh.state.SelectedPlanet)
    dynamicHeight := minimum(contentLines+6, screenHeight-4)
    modalX, modalY, modalWidth, modalHeight := meh.renderer.GetModalDimensions(screenWidth, screenHeight, dynamicHeight)

    if mouseX < modalX || mouseX >= modalX+modalWidth || mouseY < modalY || mouseY >= modalY+modalHeight {
        return false
    }

    instructionY := modalY + modalHeight - 2
    if mouseY == instructionY && len(meh.state.SelectedPlanet.Moons) > 0 {
        instruction := "Press Enter, Escape, or 'b' to close • 'm' for moons"
        mPos := strings.Index(instruction, "'m' for moons")
        if mPos >= 0 && mouseX >= modalX+2+mPos && mouseX <= modalX+2+mPos+12 {
            meh.showMoonList()
            return true
        }
    }

    if mouseY == instructionY {
        meh.state.ShowingDetails = false
        return true
    }

    return true
}

func (meh *MouseEventHandler) handlePlanetListClick(mouseX, mouseY int) bool {
    for _, pos := range meh.state.GetPlanetListPositions() {
        if mouseX >= pos.X && mouseX < pos.X+pos.Width && mouseY == pos.Y {
            meh.state.SelectedIndex = pos.Index
            meh.state.SelectedPlanet = meh.state.GetPlanets()[pos.Index]

            if !meh.state.ShowingDetails && !meh.state.ShowingMoons && !meh.state.ShowingMoonDetails && !meh.state.ShowingSystemList {
                meh.state.ShowingDetails = true
            } else if meh.state.ShowingDetails {
            }

            return true
        }
    }

    return false
}

func (meh *MouseEventHandler) showMoonDetailsInternal() {
    if meh.state.MoonSelectedIndex < len(meh.state.SelectedPlanet.Moons) {
        moonData := meh.state.SelectedPlanet.Moons[meh.state.MoonSelectedIndex]
        moonHandler := meh.renderer.GetRenderer().GetMoonHandler()
        moonName := moonHandler.GetMoonNameFromAPI(moonData)

        if moonData.ID != "" {
            if moonDetail, err := meh.planetService.GetClient().GetMoonData(moonData.ID); err == nil {
                meh.state.SelectedMoon = *moonDetail
                meh.state.SelectedMoon.BodyType = "Moon"
                meh.state.SelectedMoon.AroundPlanet = &models.Planet{
                    EnglishName: meh.state.SelectedPlanet.EnglishName,
                }
            } else {
                meh.state.SelectedMoon = models.CelestialBody{
                    ID:          moonData.ID,
                    Name:        moonData.Name,
                    EnglishName: moonName,
                    BodyType:    "Moon",
                    AroundPlanet: &models.Planet{
                        EnglishName: meh.state.SelectedPlanet.EnglishName,
                    },
                }
            }
        } else {
            meh.state.SelectedMoon = models.CelestialBody{
                ID:          moonData.ID,
                Name:        moonData.Name,
                EnglishName: moonName,
                BodyType:    "Moon",
                AroundPlanet: &models.Planet{
                    EnglishName: meh.state.SelectedPlanet.EnglishName,
                },
            }
        }

        meh.state.ShowingMoonDetails = true
        meh.state.ShowingMoons = false
    }
}
