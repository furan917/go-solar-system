# Go Solar System Explorer 🚀

Terminal app where you can explore solar syatems.

## What it does

- Navigate around planets with your keyboard & mouse
- Has some neat visualizations
- Can switch between different star systems (Solar System, Alpha Centauri, some Kepler thing, etc.)
- Real-time orbital animations for our solar system

## How to run it

```bash
git clone https://github.com/furan917/go-solar-system.git
cd go-solar-system
go mod tidy
go build
```

That should work. If it doesn't, check if you have Go installed?

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
- ✅ Multiple star systems with actual exoplanet data
- ✅ Binary system support
- ✅ Orbital animations
- ✅ Asteroid and kuiper belt represented

## What's still broken/TODO

- Some UI components need improvement
- The help system could be better (i.e could exist)
- Probably need more error handling
- Could use more star systems 

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
