package app

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/furan917/go-solar-system/internal/api"
	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/systems"
	"github.com/furan917/go-solar-system/internal/visualization"
	"github.com/gdamore/tcell/v2"
)

type SolarSystem struct {
	// Core components
	screen       tcell.Screen
	state        *AppState
	errorHandler *ErrorHandler
	logger       *log.Logger

	// Business logic components
	planetService *PlanetService
	systemManager *SystemManager

	// UI components
	renderer        *UIRenderer
	eventDispatcher *EventDispatcher
	mouseHandler    *MouseEventHandler
}

func NewSolarSystem() (*SolarSystem, error) {
	logger := log.New(os.Stderr, "[SolarSystem] ", log.LstdFlags|log.Lshortfile)

	// Initialize core dependencies
	client := api.NewClient()
	systemManager := systems.NewSystemManager("systems")
	if err := systemManager.ScanSystems(); err != nil {
		return nil, NewSystemError("failed to scan systems", err)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, NewUIError("failed to create screen", err)
	}

	if err := screen.Init(); err != nil {
		return nil, NewUIError("failed to initialize screen", err)
	}

	// Initialize state and core components
	state := NewAppState()
	errorHandler := NewErrorHandler(logger, state)
	planetService := NewPlanetService(client, systemManager)

	// Initialize rendering components
	width, height := screen.Size()
	renderer := visualization.NewRendererWithDefaults(width, height)
	uiRenderer := NewUIRenderer(screen, renderer, systemManager, state)

	// Initialize business logic components
	systemManagerComponent := NewSystemManager(state, planetService, uiRenderer, errorHandler, logger)

	// Initialize event handling components
	showMoonList := func() { state.ShowMoonList() }
	showMoonDetails := func() { /* handled by mouse handler internally */ }
	mouseHandler := NewMouseEventHandler(state, uiRenderer, showMoonList, showMoonDetails, planetService)
	eventDispatcher := NewEventDispatcher(state, mouseHandler, systemManagerComponent, planetService, uiRenderer)

	return &SolarSystem{
		screen:          screen,
		state:           state,
		errorHandler:    errorHandler,
		logger:          logger,
		planetService:   planetService,
		systemManager:   systemManagerComponent,
		renderer:        uiRenderer,
		eventDispatcher: eventDispatcher,
		mouseHandler:    mouseHandler,
	}, nil
}

func (ss *SolarSystem) Run() error {
	defer func() {
		ss.screen.Fini()
		if err := RecoverFromPanic(); err != nil {
			ss.errorHandler.HandleError(err)
		}
	}()

	// Initialize system
	if err := ss.initializeSystem(); err != nil {
		return err
	}

	// Configure screen
	ss.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	ss.screen.Clear()
	ss.screen.EnableMouse()

	// Start main loop
	return ss.runMainLoop()
}

func (ss *SolarSystem) initializeSystem() error {
	if err := ss.systemManager.LoadCurrentSystem(); err != nil {
		ss.errorHandler.HandleError(NewSystemError("failed to load initial system", err))
		return err
	}

	if err := ss.state.ValidateState(); err != nil {
		ss.errorHandler.HandleError(NewStateError("invalid state after loading", err))
	}

	if err := ss.systemManager.SortPlanetsByDistance(); err != nil {
		ss.errorHandler.HandleError(NewStateError("failed to sort planets", err))
	}

	// Process planets
	planets := ss.systemManager.NormalizePlanetNames(ss.state.GetPlanets())
	ss.state.SetPlanets(planets)

	// Add central star if needed
	centralStar := ss.systemManager.FindOrCreateCentralStar(ss.state.GetPlanets())
	if !ss.systemManager.ContainsCentralStar(ss.state.GetPlanets()) {
		ss.state.SetPlanets(append([]models.CelestialBody{centralStar}, ss.state.GetPlanets()...))
	}

	return nil
}

func (ss *SolarSystem) runMainLoop() error {
	defer func() {
		if err := RecoverFromPanic(); err != nil {
			ss.errorHandler.HandleError(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start display update goroutine
	go ss.updateDisplay(ctx)

	// Main event loop
	for ss.state.IsRunning() {
		ev := ss.screen.PollEvent()
		if err := ss.handleEventSafely(ev); err != nil {
			response := ss.errorHandler.HandleError(err)
			if response.ResetState {
				ss.state.ResetModals()
			}
			if !response.ShouldContinue {
				break
			}
		}
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (ss *SolarSystem) updateDisplay(ctx context.Context) {
	ticker := time.NewTicker(constants.DisplayUpdateRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if ss.state.IsRunning() {
				ss.renderer.DrawScreen()
			} else {
				return
			}
		}
	}
}

func (ss *SolarSystem) handleEventSafely(ev tcell.Event) error {
	defer func() {
		if r := recover(); r != nil {
			ss.logger.Printf("Panic in event handling: %v", r)
		}
	}()

	ss.eventDispatcher.HandleEvent(ev)
	return nil
}
