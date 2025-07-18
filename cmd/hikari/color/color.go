package color

import (
	"fmt"
	"math"

	"github.com/charmbracelet/lipgloss"
)

func HSBToRGB(h, s, b float64) (int, int, int) {
	s, b = s/100, b/100
	if s == 0.0 {
		return int(b * 255), int(b * 255), int(b * 255)
	}

	h = math.Mod(h, 360)
	hi := math.Floor(h / 60)
	f := h/60 - hi
	p := b * (1 - s)
	q := b * (1 - f*s)
	t := b * (1 - (1-f)*s)

	switch int(hi) {
	case 0:
		return int(b * 255), int(t * 255), int(p * 255)
	case 1:
		return int(q * 255), int(b * 255), int(p * 255)
	case 2:
		return int(p * 255), int(b * 255), int(t * 255)
	case 3:
		return int(p * 255), int(q * 255), int(b * 255)
	case 4:
		return int(t * 255), int(p * 255), int(b * 255)
	case 5:
		return int(b * 255), int(p * 255), int(q * 255)
	}

	return 0, 0, 0
}

// KelvinToRGB converts a color temperature in Kelvin to an RGB color.
// It uses a standard approximation suitable for many applications,
// but accuracy is best between 1000K and 40000K.
func KelvinToRGB(kelvin int) (r, g, b int) {
	temp := int(math.Round(float64(kelvin) / 100.0))

	// Red
	if temp <= 66 {
		r = 255
	} else {
		r = temp - 60
		r = int(329.698727446 * math.Pow(float64(r), -0.1332047592))
		r = min(max(r, 0), 255)
	}

	// Green
	if temp <= 66 {
		g = temp
		g = int(99.4708025861*math.Log(float64(g)) - 161.1195681661)
		g = min(max(g, 0), 255)
	} else {
		g = temp - 60
		g = int(288.1221695283 * math.Pow(float64(g), -0.0755148492))
		g = min(max(g, 0), 255)
	}

	// Blue
	if temp >= 66 {
		b = 255
	} else if temp <= 19 {
		b = 0
	} else {
		b = temp - 10
		b = int(138.5177312231*math.Log(float64(b)) - 305.0447927307)
		b = min(max(b, 0), 255)
	}

	return int(r), int(g), int(b)
}

func RGBToLipglossColor(r, g, b int) lipgloss.Color {
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}
