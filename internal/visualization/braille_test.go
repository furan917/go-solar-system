package visualization

import (
	"os"
	"strings"
	"testing"

	"github.com/furan917/go-solar-system/internal/models"
	"github.com/furan917/go-solar-system/internal/systems/formats"
)

func TestBrailleCanvasAllDotsInOneCell(t *testing.T) {
	canvas := NewBrailleCanvas(1, 1)
	for x := 0; x < BrailleDotsX; x++ {
		for y := 0; y < BrailleDotsY; y++ {
			canvas.Set(x, y)
		}
	}

	grid := [][]rune{{' '}}
	canvas.CompositeInto(grid)

	if grid[0][0] != '⣿' {
		t.Errorf("got U+%04X, want U+28FF", grid[0][0])
	}
}

func TestBrailleCanvasDotMapping(t *testing.T) {
	cases := []struct {
		subX, subY int
		want       rune
	}{
		{0, 0, '⠁'},
		{0, 1, '⠂'},
		{0, 2, '⠄'},
		{0, 3, '⡀'},
		{1, 0, '⠈'},
		{1, 1, '⠐'},
		{1, 2, '⠠'},
		{1, 3, '⢀'},
	}

	for _, c := range cases {
		canvas := NewBrailleCanvas(1, 1)
		canvas.Set(c.subX, c.subY)

		grid := [][]rune{{' '}}
		canvas.CompositeInto(grid)

		if grid[0][0] != c.want {
			t.Errorf("dot (%d,%d): got U+%04X, want U+%04X", c.subX, c.subY, grid[0][0], c.want)
		}
	}
}

func TestBrailleCanvasIgnoresOutOfBounds(t *testing.T) {
	canvas := NewBrailleCanvas(2, 2)

	canvas.Set(-1, 0)
	canvas.Set(0, -1)
	canvas.Set(2*BrailleDotsX, 0)
	canvas.Set(0, 2*BrailleDotsY)

	grid := [][]rune{{' ', ' '}, {' ', ' '}}
	canvas.CompositeInto(grid)

	for y := range grid {
		for x := range grid[y] {
			if grid[y][x] != ' ' {
				t.Errorf("cell %d,%d was written: U+%04X", x, y, grid[y][x])
			}
		}
	}
}

func TestBrailleCanvasDoesNotOverwriteExistingGlyphs(t *testing.T) {
	canvas := NewBrailleCanvas(1, 1)
	canvas.Set(0, 0)

	grid := [][]rune{{'♁'}}
	canvas.CompositeInto(grid)

	if grid[0][0] != '♁' {
		t.Errorf("planet glyph was overwritten by U+%04X", grid[0][0])
	}
}

func TestBrailleCanvasHandlesShortGrid(t *testing.T) {
	canvas := NewBrailleCanvas(4, 4)
	canvas.Set(0, 0)
	canvas.Set(6, 6)

	grid := [][]rune{{' '}}
	canvas.CompositeInto(grid)

	if grid[0][0] == ' ' {
		t.Error("expected the in-range cell to be written")
	}
}

func TestSimpleOrbitsRequested(t *testing.T) {
	cases := map[string]bool{
		"":      false,
		"0":     false,
		"false": false,
		"no":    false,
		"off":   false,
		"OFF":   false,
		"1":     true,
		"true":  true,
		"yes":   true,
		" 1 ":   true,
	}

	for value, want := range cases {
		t.Setenv(SimpleOrbitsEnvVar, value)
		if got := SimpleOrbitsRequested(); got != want {
			t.Errorf("%q: got %v, want %v", value, got, want)
		}
	}
}

func TestIsBraille(t *testing.T) {
	if !IsBraille('⠀') || !IsBraille('⣿') {
		t.Error("braille range endpoints not recognised")
	}
	if IsBraille('·') || IsBraille('⟿') || IsBraille('⤀') {
		t.Error("non-braille rune recognised as braille")
	}
}

func renderTestSystem(t *testing.T, width, height int) [][]rune {
	t.Helper()

	data, err := os.ReadFile("../../systems/toi-178.json")
	if err != nil {
		t.Fatal(err)
	}
	sys, err := formats.NewJSONFormat().ParseSystemData(data)
	if err != nil {
		t.Fatal(err)
	}

	return NewRendererWithDefaults(width, height).RenderSolarSystemData(sys.Bodies, width, height)
}

func countRunes(grid [][]rune) (braille int, dots int) {
	for _, row := range grid {
		for _, r := range row {
			if IsBraille(r) {
				braille++
			}
			if r == '·' {
				dots++
			}
		}
	}
	return
}

func TestOrbitsRenderAsBrailleByDefault(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "")

	braille, dots := countRunes(renderTestSystem(t, 176, 32))

	if braille == 0 {
		t.Error("expected braille orbit cells")
	}
	if dots != 0 {
		t.Errorf("expected no '·' orbit cells, got %d", dots)
	}
}

func TestSimpleOrbitsEnvRestoresPlainDots(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "1")

	braille, dots := countRunes(renderTestSystem(t, 176, 32))

	if dots == 0 {
		t.Error("expected '·' orbit cells when simple orbits are requested")
	}
	if braille != 0 {
		t.Errorf("expected no braille cells, got %d", braille)
	}
}

func TestBrailleOrbitsCoverMoreCellsThanPlainDots(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "")
	braille, _ := countRunes(renderTestSystem(t, 176, 32))

	t.Setenv(SimpleOrbitsEnvVar, "1")
	_, dots := countRunes(renderTestSystem(t, 176, 32))

	if braille <= dots {
		t.Errorf("braille orbits (%d cells) should be denser than plain dots (%d cells)", braille, dots)
	}
}

func TestPlanetGlyphsSurviveBrailleCompositing(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "")

	grid := renderTestSystem(t, 176, 32)

	var found bool
	for _, row := range grid {
		if strings.ContainsRune(string(row), '✩') {
			found = true
			break
		}
	}

	if !found {
		t.Error("star glyphs were lost when braille orbits were composited")
	}
}

func loadTestSystem(t testing.TB, file string) []models.CelestialBody {
	t.Helper()

	data, err := os.ReadFile("../../systems/" + file)
	if err != nil {
		t.Fatal(err)
	}
	sys, err := formats.NewJSONFormat().ParseSystemData(data)
	if err != nil {
		t.Fatal(err)
	}

	return sys.Bodies
}

func TestOrbitsRenderAsBrailleOnThePathTheAppUses(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "")

	const w, h = 176, 32
	grid, positions := NewRendererWithDefaults(w, h).
		RenderSolarSystemDataWithPositions(loadTestSystem(t, "toi-178.json"), w, h, 180, 40)

	braille, dots := countRunes(grid)

	if braille == 0 {
		t.Error("expected braille orbit cells on the rendering path the app uses")
	}
	if dots != 0 {
		t.Errorf("expected no '·' orbit cells, got %d", dots)
	}
	if len(positions) == 0 {
		t.Error("expected planet positions to still be reported")
	}
}

func TestSimpleOrbitsEnvAppliesOnThePathTheAppUses(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "1")

	const w, h = 176, 32
	grid, _ := NewRendererWithDefaults(w, h).
		RenderSolarSystemDataWithPositions(loadTestSystem(t, "toi-178.json"), w, h, 180, 40)

	braille, dots := countRunes(grid)

	if dots == 0 {
		t.Error("expected '·' orbit cells when simple orbits are requested")
	}
	if braille != 0 {
		t.Errorf("expected no braille cells, got %d", braille)
	}
}

func BenchmarkRenderWithPositionsBraille(b *testing.B) {
	b.Setenv(SimpleOrbitsEnvVar, "")
	benchmarkRender(b)
}

func BenchmarkRenderWithPositionsSimple(b *testing.B) {
	b.Setenv(SimpleOrbitsEnvVar, "1")
	benchmarkRender(b)
}

func benchmarkRender(b *testing.B) {
	const w, h = 176, 32
	bodies := loadTestSystem(b, "kepler-90.json")
	r := NewRendererWithDefaults(w, h)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderSolarSystemDataWithPositions(bodies, w, h, 180, 40)
	}
}

func ringMidX(grid [][]rune, match func(rune) bool) (float64, int) {
	minX, maxX, n := 1<<30, -1, 0
	for _, row := range grid {
		for x, r := range row {
			if !match(r) {
				continue
			}
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	return float64(minX+maxX) / 2, n
}

func TestRingsShareTheGridCentre(t *testing.T) {
	t.Setenv(SimpleOrbitsEnvVar, "")

	const w, h = 176, 32
	grid, _ := NewRendererWithDefaults(w, h).
		RenderSolarSystemDataWithPositions(loadTestSystem(t, "alpha-centauri.json"), w, h, 180, 40)

	rings := map[string]func(rune) bool{
		"orbit":    IsBraille,
		"asteroid": func(r rune) bool { return r == '∗' },
		"kuiper":   func(r rune) bool { return r == '◦' },
	}

	for name, match := range rings {
		mid, n := ringMidX(grid, match)
		if n == 0 {
			t.Fatalf("%s ring not rendered", name)
		}
		if mid != float64(w/2) {
			t.Errorf("%s ring centred on x=%.1f, want %d", name, mid, w/2)
		}
	}
}
