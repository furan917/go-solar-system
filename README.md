# Go Solar System Explorer 🚀

Yeah, so this is basically a terminal app where you can explore space stuff. It's pretty cool I guess. Still working on it though.

## What it does

- Navigate around planets with your keyboard (because who needs a mouse in space?)
- Shows you actual NASA data about planets and their moons
- Has some neat visualizations that look surprisingly good in a terminal
- Can switch between different star systems (Solar System, Alpha Centauri, some Kepler thing, etc.)
- Real-time orbital animations because why not

## How to run it

```bash
git clone https://github.com/furan917/go-solar-system.git
cd go-solar-system
go mod tidy
go run main.go
```

That should work. If it doesn't, uh... check if you have Go installed?

## Controls (the important stuff)

**Basic navigation:**
- Arrow keys = move around
- Enter = see planet details
- Numbers 1-9 = jump to specific planets/sun
- S = switch between star systems
- Q = quit (or Escape, whatever)

**When looking at planet details:**
- M = view moons (if the planet has any)
- B = go back
- Q = still quits

**Moon stuff:**
- Up/Down = navigate moon list
- Enter = moon details
- Escape/B = back to planet

## Current features (aka what actually works)

- ✅ All the basic planet browsing stuff
- ✅ Real NASA/JPL data (pretty accurate)
- ✅ Moon exploration (with way too many of Jupiter's moons)
- ✅ Multiple star systems with actual exoplanet data
- ✅ Binary and triple star systems (they actually orbit each other!)
- ✅ Mouse clicking on planets (surprisingly satisfying)
- ✅ Orbital animations that don't make your terminal explode
- ✅ Asteroid belts because why not

## What's still broken/TODO

- Some UI components are being refactored (it's a work in progress, okay?)
- The help system could be better
- Probably need more error handling
- Could use more star systems
- Maybe add export functionality someday

## Architecture (for the nerds)

```
├── main.go                    # Entry point
├── internal/
│   ├── models/               # Planet/moon data structures
│   ├── api/                  # Talks to NASA APIs
│   ├── visualization/        # Makes pretty terminal graphics
│   ├── systems/              # Star system management
│   ├── ui/                   # Modal components (new hotness)
│   └── app/                  # Main app logic (some legacy stuff)
```

The code follows some clean architecture principles, mostly. There's a bit of legacy code hanging around that I'm slowly cleaning up.

## Testing

```bash
go test ./...
```

Tests exist and they pass (usually).

## Data sources

Uses real data from:
- [Solar System OpenData API](https://api.le-systeme-solaire.net/en/)
- NASA/JPL databases
- Various exoplanet catalogs

## Star systems included

- **Solar System**: Our neighborhood (obviously)
- **Alpha Centauri**: Closest star system with some interesting planets
- **Kepler-452**: Has "Earth's cousin" planet
- **TRAPPIST-1**: 7 Earth-sized planets, pretty cool

## Contributing

Sure, if you want to help out:

1. Fork it
2. Make changes
3. Don't break anything
4. Submit a PR

Standard stuff.

## License

MIT License - do whatever you want with it.

---

*Built with Go and still a work in progress but it's getting there.*