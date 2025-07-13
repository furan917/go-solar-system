package visualization

import "math"

// CircleDrawer handles drawing circular shapes with proper aspect ratio compensation
type CircleDrawer struct {
	aspectRatio float64
}

// NewCircleDrawer creates a new circle drawer with the specified aspect ratio
func NewCircleDrawer(aspectRatio float64) *CircleDrawer {
	return &CircleDrawer{
		aspectRatio: aspectRatio,
	}
}

// DrawCircle draws a circle outline on the grid with improved algorithm
func (cd *CircleDrawer) DrawCircle(grid [][]rune, centerX, centerY int, radius float64, symbol rune) {
	circumference := 2 * math.Pi * radius
	steps := int(circumference * 4)
	if steps < 720 {
		steps = 720
	}

	for i := 0; i < steps; i++ {
		angle := float64(i) * 2 * math.Pi / float64(steps)
		x := centerX + int(radius*math.Cos(angle)*cd.aspectRatio)
		y := centerY + int(radius*math.Sin(angle))

		if cd.isInBounds(x, y, len(grid[0]), len(grid)) && grid[y][x] == ' ' {
			grid[y][x] = symbol
		}
	}
}

// DrawFilledCircle draws a filled circle on the grid
func (cd *CircleDrawer) DrawFilledCircle(grid [][]rune, centerX, centerY, radius int, symbol rune) {
	for dy := -radius; dy <= radius; dy++ {
		rowWidth := math.Sqrt(float64(radius*radius - dy*dy))
		maxDx := int(rowWidth * cd.aspectRatio)

		for dx := -maxDx; dx <= maxDx; dx++ {
			x := centerX + dx
			y := centerY + dy

			if cd.isInBounds(x, y, len(grid[0]), len(grid)) {
				grid[y][x] = symbol
			}
		}
	}
}

// CalculatePosition calculates a position on a circle at the given angle
func (cd *CircleDrawer) CalculatePosition(centerX, centerY int, radius float64, angle float64) (int, int) {
	x := centerX + int(radius*math.Cos(angle)*cd.aspectRatio)
	y := centerY + int(radius*math.Sin(angle))
	return x, y
}

// isInBounds checks if coordinates are within grid bounds
func (cd *CircleDrawer) isInBounds(x, y, width, height int) bool {
	return x >= 0 && x < width && y >= 0 && y < height
}
