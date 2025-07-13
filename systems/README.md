# External Star Systems

This directory contains JSON files defining external star systems that can be explored in the Solar System Explorer application.

## Adding New Star Systems

### File Format

Create a new `.json` file in this directory with the following structure:

```json
{
  "systemName": "Your System Name",
  "description": "Brief description of the system",
  "discoveryYear": "YYYY or YYYY-YYYY for range",
  "distance": "X.X light-years or parsecs",
  "bodies": [
    {
      "id": "unique-body-id",
      "name": "Body Name",
      "englishName": "Body Name",
      "bodyType": "Star" | "Planet",
      "isPlanet": false,
      "meanRadius": 695700,
      "mass": {
        "massValue": 1.989,
        "massExponent": 30
      },
      "semimajorAxis": 0,
      "sideralOrbit": 365.25,
      "sideralRotation": 24.0,
      "density": 1.408,
      "gravity": 274.0,
      "eccentricity": 0.0167,
      "inclination": 0.0,
      "discoveredBy": "Discovery Method/Team",
      "discoveryDate": "YYYY",
      "moons": [],
      "temperature": 5778,
      "stellarClass": "G2V",
      "age": 4600000000
    }
  ]
}
```

### Required Fields

#### System Metadata
- **systemName**: Display name for the system
- **description**: Brief description (shown in system selection)
- **discoveryYear**: When the system was discovered
- **distance**: Distance from Earth with units

#### Star Bodies (`bodyType: "Star"`)
- **id**: Unique identifier
- **name** & **englishName**: Display names
- **bodyType**: Must be "Star"
- **isPlanet**: Must be `false`
- **meanRadius**: Radius in kilometers
- **mass**: Object with `massValue` and `massExponent` (scientific notation)
- **semimajorAxis**: Must be `0` for stars
- **temperature**: Effective temperature in Kelvin
- **stellarClass**: Stellar classification (see Stellar Classification below)

#### Planet Bodies (`bodyType: "Planet"`)
- **id**: Unique identifier
- **name** & **englishName**: Display names
- **bodyType**: Must be "Planet"
- **isPlanet**: Must be `true`
- **meanRadius**: Radius in kilometers
- **mass**: Object with `massValue` and `massExponent`
- **semimajorAxis**: Distance from star(s) in kilometers
- **sideralOrbit**: Orbital period in days
- **eccentricity**: Orbital eccentricity (0-1)
- **inclination**: Orbital inclination in degrees

### Stellar Classification

The application automatically assigns star symbols based on stellar class:

| Stellar Class | Symbol | Color  | Description               |
|---------------|--------|--------|---------------------------|
| **O, B**      | ✦      | Blue   | Hot blue/blue-white stars |
| **A, F**      | ✧      | White  | White/yellow-white stars  |
| **G**         | ☉      | Yellow | Yellow stars (Sun-like)   |
| **K**         | ✩      | Orange | Orange stars              |
| **M**         | ✪      | Red    | Red dwarf stars           |

**Examples:**
- G2V: Sun-like yellow dwarf
- M8V: Cool red dwarf  
- K1V: Orange dwarf
- A5V: White main sequence
- B5V: Blue-white giant

### Multi-Star Systems

#### Binary Stars
For binary star systems, include two star bodies with different masses. The application will automatically:
- Calculate barycenter positioning based on mass ratios
- Apply realistic orbital periods using Kepler's laws
- Render animated orbital motion

#### Triple+ Star Systems
For systems with 3+ stars, they will be arranged in a rotating ring formation around the system barycenter.

### Optional Fields

#### All Bodies
- **sideralRotation**: Rotation period in hours
- **density**: Density in g/cm³
- **gravity**: Surface gravity in m/s²
- **discoveredBy**: Discovery method or team
- **discoveryDate**: Discovery year
- **moons**: Array of moon objects (for planets)

#### Planets Only
- **equilibriumTemperature**: Temperature in Kelvin
- **habitableZone**: Boolean indicating if in habitable zone
- **escapeVelocity**: Escape velocity in km/s

#### Stars Only
- **age**: Age in years

### Moon Format

For planets with moons:

```json
"moons": [
  {
    "id": "moon-id",
    "name": "Moon Name",
    "englishName": "Moon Name"
  }
]
```

## Real Data Sources

When creating systems, use real astronomical data from:

- **NASA Exoplanet Archive**: https://exoplanetarchive.ipac.caltech.edu/
- **SIMBAD Database**: http://simbad.u-strasbg.fr/simbad/
- **Exoplanet Catalog**: http://exoplanet.eu/
- **NASA/JPL Solar System Dynamics**: https://ssd.jpl.nasa.gov/

## Example Systems

### Binary System (Alpha Centauri)
- Two stars with different masses and stellar classes
- Planets orbiting around the system barycenter
- Realistic stellar properties and orbital mechanics

### Single Star System (TRAPPIST-1)
- Ultra-cool M-dwarf star
- Multiple terrestrial planets
- Some in the habitable zone

### Sun-like System (Kepler-452)
- G-type star similar to our Sun
- Mix of terrestrial and gas giant planets
- Earth-analog planet in habitable zone

## Testing Your System

1. Add your JSON file to the `systems/` directory
2. Build and run the application: `go run main.go`
3. Press 'S' to open system selection
4. Your system should appear in the alphabetically sorted list
5. Select it to verify:
   - Stars render with correct symbols and colors
   - Planets orbit appropriately
   - Modal information displays correctly
   - Multi-star orbital dynamics work (if applicable)

## System Validation

The application will automatically validate:
- ✅ JSON syntax correctness
- ✅ Required fields present
- ✅ Star vs planet classification
- ✅ Mass and radius values are positive
- ✅ Orbital parameters are reasonable

Common issues:
- ❌ Missing `bodyType` field
- ❌ Stars with `isPlanet: true`
- ❌ Planets with `semimajorAxis: 0`
- ❌ Invalid stellar classification
- ❌ Negative mass or radius values

## Performance Considerations

- System metadata (name, description, etc.) is cached for fast list display
- Full system data is loaded only when switching to that system
- Keep systems focused on key bodies to maintain performance
- Large numbers of moons may impact rendering performance