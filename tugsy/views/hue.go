package views

const (
	UnknownHue = Hue(361)
	UnknownR   = 128
	UnknownG   = 128
	UnknownB   = 128
)

// Returns the RGB values for the given hue, assuming saturation and value
// equal to 1.0. This is a simplification of the general formula, with C
// equal to 1 and m equal to 0.
//func computeRGB(hue Hue) (r, g, b uint8) {
//
//	// X = C × (1 - |(H / 60°) mod 2 - 1|)
//	x := 1 - math.Abs(hue/60.0 % 2 - 1)
//	X := uint8(x * 255 + 0.5)
//
//	switch {
//	case hue < 60:
//		return 255, X, 0
//	case hue < 120:
//		return X, 255, 0
//	case hue < 180:
//		return 0, 255, X
//	case hue < 240:
//		return 0, X, 255
//	case hue < 300:
//		return X, 0, 255
//	case hue <= 360:
//		return 255, 0, X
//	default:
//		logger.Warn("Got an invalid hue value", "hue", hue)
//		return 128, 128, 128
//	}
//}

// Maps the given Hue value to an RGB triplet, returning neutral gray if
// Hue == 361 (an ordinarily invalid value)
func HueToRGB(hue Hue) (r, g, b uint8) {
	switch hue {
	case 10:
		return 255, 43, 0
	case 30:
		return 255, 128, 0
	case 50:
		return 255, 212, 0
	case 70:
		return 212, 255, 0
	case 90:
		return 128, 255, 0
	case 110:
		return 43, 255, 0
	case 130:
		return 0, 255, 43
	case 150:
		return 0, 255, 128
	case 170:
		return 0, 255, 212
	case 190:
		return 0, 212, 255
	case 210:
		return 0, 128, 255
	case 230:
		return 0, 43, 255
	case 250:
		return 43, 0, 255
	case 270:
		return 128, 0, 255
	case 290:
		return 212, 0, 255
	case 310:
		return 255, 0, 212
	case 330:
		return 255, 0, 128
	case 350:
		return 255, 0, 43
	case UnknownHue:
		return UnknownR, UnknownG, UnknownB
	default:
		logger.Warn("Got an invalid hue value", "hue", hue)
		return UnknownR, UnknownG, UnknownB
	}
}
