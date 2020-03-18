package views

import (
	"math"
	"reflect"
	"sync"

	"github.com/joemadeus/tugsy/tugsy/shipdata"
	logger "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultDestSpriteSizePixels = 20
)

// ShipInfoElement renders information for a ship, including its registration, flag,
// name, current destination and current situation (moored, underway, etc) into a
// BaseInfoElement
type ShipInfoElement struct {
	*SpriteSet

	history *shipdata.ShipHistory
}

func NewShipInfoElement(sprites *SpriteSet, h *shipdata.ShipHistory) *ShipInfoElement {
	return &ShipInfoElement{SpriteSet: sprites, history: h}
}

func (e *ShipInfoElement) ClosestChild(x, y int32) (ChildElement, float64) {
	return nil, math.MaxFloat64
}

func (e *ShipInfoElement) Render(v *View) error {
	tex, err := v.ScreenRenderer.CreateTexture(PixelFormat, sdl.TEXTUREACCESS_STATIC, 0, 0)
	if err != nil {
		return err
	}

	flag, err := e.SpriteSet.FlagSheet.GetSprite("JO")
	if err != nil {
		return err
	}

	return nil
}

type AllPositionElements struct {
	sync.Mutex
	*SpriteSet

	aisData          *shipdata.AISData
	positionElements map[uint32]*ShipPositionElement
	baseInfoElement  *BaseInfoElement
}

func NewAllPositionElements(sprites *SpriteSet, ais *shipdata.AISData, be *BaseInfoElement) *AllPositionElements {
	return &AllPositionElements{
		SpriteSet:        sprites,
		aisData:          ais,
		positionElements: make(map[uint32]*ShipPositionElement),
		baseInfoElement:  be,
	}
}

func (e *AllPositionElements) ClosestChild(x, y int32) (ChildElement, float64) {
	e.Lock()
	defer e.Unlock()

	closest := struct {
		ele *ShipPositionElement
		d   float64
	}{d: math.MaxFloat64}
	for _, sp := range e.positionElements {
		d := sp.Distance(x, y)
		if d > closest.d {
			continue
		}

		closest.d = d
		closest.ele = sp
	}

	logger.Debugf("AllPositionElements ClosestChild at %s, %f", reflect.TypeOf(closest.ele).String(), closest.d)
	return closest.ele, closest.d
}

func (e *AllPositionElements) Render(v *View) error {
	mmsis := make(map[uint32]struct{})
	histories := e.aisData.ShipHistories() // returns a copy

	e.Lock()
	defer e.Unlock()

	for _, sh := range histories {
		se, ok := e.positionElements[sh.MMSI]
		if ok == false {
			se = &ShipPositionElement{SpriteSet: e.SpriteSet, history: sh, baseInfoElement: e.baseInfoElement}
			e.positionElements[sh.MMSI] = se
		}

		if err := se.Render(v); err != nil {
			return err
		}

		mmsis[sh.MMSI] = struct{}{}
	}

	// prune ShipPositionElements that no longer exist. could be replaced, along
	// with add(), with a chan, I suppose
	for m := range e.positionElements {
		if _, ok := mmsis[m]; ok == false {
			delete(e.positionElements, m)
		}
	}

	return nil
}

type ShipPositionElement struct {
	*SpriteSet

	curPosition     BaseMapPosition
	history         *shipdata.ShipHistory
	baseInfoElement *BaseInfoElement
}

func (e *ShipPositionElement) Distance(x, y int32) float64 {
	d := screenDistance(x, y, e.curPosition.X, e.curPosition.Y)
	logger.Debugf("ShipPositionElement distance is %f", d)
	return d
}

func (e *ShipPositionElement) HandleTouch() error {
	logger.Debug("Handling touch in ShipPositionElement")
	infoElement := NewShipInfoElement(e.SpriteSet, e.history)
	return e.baseInfoElement.UpdateContent(infoElement)
}

func (e *ShipPositionElement) Render(v *View) error {
	positions := e.history.Positions() // copy
	if len(positions) == 0 {
		return nil
	}

	// TODO we're reloading sprites and primitives every time through. cut that
	//  out and start holding some view state

	hue := shipTypeToHue(e.history)
	if err := e.renderHistory(v, hue, positions); err != nil {
		return err
	}

	if err := e.renderPosition(v, hue, positions); err != nil {
		return err
	}

	return nil
}

func (e *ShipPositionElement) renderPosition(view *View, hue Hue, positions []shipdata.Positionable) error {
	e.curPosition = view.BaseMapPosition(positions[len(positions)-1].GetPositionReport())

	var sprite *Sprite
	var err error
	if hue == UnknownHue {
		// return the special "unknown" dot
		if sprite, err = e.SpecialSheet.GetSprite("unknown"); err != nil {
			logger.WithError(err).Error("could not load special sprite 'unknown'")
			return err
		}
	} else {
		if sprite, err = e.DotSheet.GetSprite(hue, "normal"); err != nil {
			logger.WithError(err).Errorf("could not load sprite with hue %v", hue)
			return err
		}
	}

	// TODO: Set opacity to 33% if older than a certain age

	// TODO: Set hazardous cargo markers

	if err := view.ScreenRenderer.Copy(sprite.Texture, sprite.Rect, toDestRect(&e.curPosition, defaultDestSpriteSizePixels)); err != nil {
		logger.WithError(err).Error("rendering ship history")
		return err
	}

	return nil
}

func (e *ShipPositionElement) renderHistory(view *View, hue Hue, positions []shipdata.Positionable) error {
	trackPointsSize := int32(4)
	trackAlpha := uint8(128)

	r, g, b := HueToRGB(hue)
	sdlPoints := make([]sdl.Point, len(positions), len(positions))
	sdlRects := make([]sdl.Rect, len(positions), len(positions))

	for i, position := range positions {
		baseMapPosition := view.BaseMapPosition(position.GetPositionReport())
		sdlPoints[i] = sdl.Point{
			X: int32(baseMapPosition.X + 0.5),
			Y: int32(baseMapPosition.Y + 0.5),
		}
		sdlRects[i] = sdl.Rect{
			X: int32(baseMapPosition.X+0.5) - (trackPointsSize / 2),
			Y: int32(baseMapPosition.Y+0.5) - (trackPointsSize / 2),
			W: trackPointsSize,
			H: trackPointsSize,
		}
	}

	if err := view.ScreenRenderer.SetDrawColor(r, g, b, trackAlpha); err != nil {
		logger.WithError(err).Warn("setting the draw color")
		return err
	}

	if err := view.ScreenRenderer.DrawLines(sdlPoints); err != nil {
		logger.WithError(err).Warn("rendering track lines")
		return err
	}

	if err := view.ScreenRenderer.DrawRects(sdlRects); err != nil {
		logger.WithError(err).Warn("rendering track points")
		return err
	}

	return nil
}

func toDestRect(position *BaseMapPosition, pixSquare int32) *sdl.Rect {
	return &sdl.Rect{
		X: int32(position.X+0.5) - (pixSquare / 2),
		Y: int32(position.Y+0.5) - (pixSquare / 2),
		W: pixSquare,
		H: pixSquare,
	}
}

// Maps a ship type to a hue, or to UnknownHue if the type is unknown or should be
// mapped that way anyway
func shipTypeToHue(history *shipdata.ShipHistory) Hue {
	// TODO could easily be config
	voyagedata := history.VoyageData()
	switch {
	case voyagedata == nil:
		return UnknownHue
	case voyagedata.ShipType <= 29:
		return UnknownHue
	case voyagedata.ShipType == 30:
		// fishing
	case voyagedata.ShipType <= 32:
		// towing -- VIOLET: H310
		return 310
	case voyagedata.ShipType <= 34:
		// diving/dredging/underwater
	case voyagedata.ShipType == 35:
		// military ops
	case voyagedata.ShipType == 36:
		// sailing
	case voyagedata.ShipType == 37:
		// pleasure craft -- VIOLET: H290
		return 290
	case voyagedata.ShipType <= 39:
		return UnknownHue
	case voyagedata.ShipType <= 49:
		// high speed craft -- YELLOW/ORANGE: H50
		return 50
	case voyagedata.ShipType == 50:
		// pilot vessel -- ORANGE: H30
		return 30
	case voyagedata.ShipType == 51:
		// search & rescue
	case voyagedata.ShipType == 52:
		// tug -- RED: H10
		return 10
	case voyagedata.ShipType == 53:
		// port tender -- ORANGE: H50
		return 50
	case voyagedata.ShipType == 54:
		return UnknownHue // "anti pollution equipment"
	case voyagedata.ShipType == 55:
		// law enforcement
	case voyagedata.ShipType <= 57:
		return UnknownHue
	case voyagedata.ShipType == 58:
		// medical transport
	case voyagedata.ShipType == 59:
		// "noncombatant ship"
	case voyagedata.ShipType <= 69:
		// passenger -- GREEN: H110
		return 110
	case voyagedata.ShipType <= 79:
		// cargo -- LIGHT BLUE: H190
		return 190
	case voyagedata.ShipType <= 89:
		// tanker -- DARK BLUE: H250
		return 250
	case voyagedata.ShipType <= 99:
		return UnknownHue // other
	}

	logger.WithField("type num", voyagedata.ShipType).Warn("mapping an unhandled ship type")
	return 0
}

var MIDInformalNames = map[int]string{
	// TODO could easily be config
	201: "Albania",
	202: "Andorra",
	203: "Austria",
	204: "Azores (PT)",
	205: "Belgium",
	206: "Belarus",
	207: "Bulgaria",
	208: "Vatican City State",
	209: "Cyprus",
	210: "Cyprus",
	211: "Germany",
	212: "Cyprus",
	213: "Georgia",
	214: "Moldova",
	215: "Malta",
	216: "Armenia",
	218: "Germany",
	219: "Denmark",
	220: "Denmark",
	224: "Spain",
	225: "Spain",
	226: "France",
	227: "France",
	228: "France",
	229: "Malta",
	230: "Finland",
	231: "Faroe Islands (DK)",
	232: "United Kingdom",
	233: "United Kingdom",
	234: "United Kingdom",
	235: "United Kingdom",
	236: "Gibraltar (UK)",
	237: "Greece",
	238: "Croatia",
	239: "Greece",
	240: "Greece",
	241: "Greece",
	242: "Morocco",
	243: "Hungary",
	244: "Netherlands",
	245: "Netherlands",
	246: "Netherlands",
	247: "Italy",
	248: "Malta",
	249: "Malta",
	250: "Ireland",
	251: "Iceland",
	252: "Liechtenstein",
	253: "Luxembourg",
	254: "Monaco",
	255: "Madeira (PT)",
	256: "Malta",
	257: "Norway",
	258: "Norway",
	259: "Norway",
	261: "Poland",
	262: "Montenegro",
	263: "Portugal",
	264: "Romania",
	265: "Sweden",
	266: "Sweden",
	267: "Slovak Republic",
	268: "San Marino",
	269: "Switzerland",
	270: "Czech Republic",
	271: "Turkey",
	272: "Ukraine",
	273: "Russian Federation",
	274: "Macedonia",
	275: "Latvia",
	276: "Estonia",
	277: "Lithuania",
	278: "Slovenia",
	279: "Serbia",
	301: "Anguilla (UK)",
	303: "Alaska (US)",
	304: "Antigua & Barbuda",
	305: "Antigua & Barbuda",
	306: "NL Caribbean Islands",
	307: "Aruba (NL)",
	308: "Bahamas",
	309: "Bahamas",
	310: "Bermuda (UK)",
	311: "Bahamas",
	312: "Belize",
	314: "Barbados",
	316: "Canada",
	319: "Cayman Islands (UK)",
	321: "Costa Rica",
	323: "Cuba",
	325: "Dominica",
	327: "Dominican Republic",
	329: "Guadeloupe (FR)",
	330: "Grenada",
	331: "Greenland (DK)",
	332: "Guatemala",
	334: "Honduras",
	336: "Haiti",
	338: "United States",
	339: "Jamaica",
	341: "St. Kitts & Nevis",
	343: "St. Lucia",
	345: "Mexico",
	347: "Martinique (FR)",
	348: "Montserrat (UK)",
	350: "Nicaragua",
	351: "Panama",
	352: "Panama",
	353: "Panama",
	354: "Panama",
	355: " - ",
	356: " - ",
	357: " - ",
	358: "Puerto Rico (US)",
	359: "El Salvador",
	361: "St. Pierre & Miquelon (FR)",
	362: "Trinidad & Tobago",
	364: "Turks & Caicos (UK)",
	366: "United States",
	367: "United States",
	368: "United States",
	369: "United States",
	370: "Panama",
	371: "Panama",
	372: "Panama",
	373: "Panama",
	375: "St. Vincent & the Grenadines",
	376: "St. Vincent & the Grenadines",
	377: "St. Vincent & the Grenadines",
	378: "Virgin Islands (UK)",
	379: "Virgin Islands (US)",
	401: "Afghanistan",
	403: "Saudi Arabia",
	405: "Bangladesh",
	408: "Bahrain",
	410: "Bhutan",
	412: "China",
	413: "China",
	414: "China",
	416: "Taiwan (CN)",
	417: "Sri Lanka",
	419: "India",
	422: "Iran",
	423: "Azerbaijan",
	425: "Iraq",
	428: "Israel",
	431: "Japan",
	432: "Japan",
	434: "Turkmenistan",
	436: "Kazakhstan",
	437: "Uzbekistan",
	438: "Jordan",
	440: "Korea",
	441: "Korea",
	443: "State of Palestine",
	445: "Korea (DPR)",
	447: "Kuwait",
	450: "Lebanon",
	451: "Kyrgyz Republic",
	453: "Macao (CN)",
	455: "Maldives",
	457: "Mongolia",
	459: "Nepal",
	461: "Oman",
	463: "Pakistan",
	466: "Qatar",
	468: "Syria",
	470: "United Arab Emirates",
	472: "Tajikistan",
	473: "Yemen",
	475: "Yemen",
	477: "Hong Kong (CN)",
	478: "Bosnia & Herzegovina",
	501: "Adelie Land (FR)",
	503: "Australia",
	506: "Myanmar",
	508: "Brunei Darussalam",
	510: "Micronesia",
	511: "Palau",
	512: "New Zealand",
	514: "Cambodia",
	515: "Cambodia",
	516: "Christmas Island (AU)",
	518: "Cook Islands (NZ)",
	520: "Fiji",
	523: "Cocos Islands (AU)",
	525: "Indonesia",
	529: "Kiribati",
	531: "Laos",
	533: "Malaysia",
	536: "Northern Mariana Islands (US)",
	538: "Marshall Islands",
	540: "New Caledonia (FR)",
	542: "Niue (NZ)",
	544: "Nauru",
	546: "French Polynesia (FR)",
	548: "Philippines",
	553: "Papua New Guinea",
	555: "Pitcairn Island (UK)",
	557: "Solomon Islands",
	559: "American Samoa (US)",
	561: "Samoa",
	563: "Singapore",
	564: "Singapore",
	565: "Singapore",
	566: "Singapore",
	567: "Thailand",
	570: "Tonga",
	572: "Tuvalu",
	574: "Viet Nam",
	576: "Vanuatu",
	577: "Vanuatu",
	578: "Wallis & Futuna Islands (FR)",
	601: "South Africa",
	603: "Angola",
	605: "Algeria",
	607: "St. Paul, Amsterdam Islands (FR)",
	608: "Ascension Island (UK)",
	609: "Burundi",
	610: "Benin",
	611: "Botswana",
	612: "Central African Republic",
	613: "Cameroon",
	615: "Congo",
	616: "Comoros",
	617: "Cabo Verde",
	618: "Crozet Archipelago (FR)",
	619: "Cote d'Ivoire",
	620: "Comoros",
	621: "Djibouti",
	622: "Egypt",
	624: "Ethiopia",
	625: "Eritrea",
	626: "Gabonese Republic",
	627: "Ghana",
	629: "Gambia",
	630: "Guinea-Bissau",
	631: "Equatorial Guinea",
	632: "Guinea",
	633: "Burkina Faso",
	634: "Kenya",
	635: "Kerguelen Islands (FR)",
	636: "Liberia",
	637: "Liberia",
	638: "South Sudan",
	642: "Libya",
	644: "Lesotho",
	645: "Mauritius",
	647: "Madagascar",
	649: "Mali",
	650: "Mozambique",
	654: "Mauritania",
	655: "Malawi",
	656: "Niger",
	657: "Nigeria",
	659: "Namibia",
	660: "Reunion (FR)",
	661: "Rwanda",
	662: "Sudan",
	663: "Senegal",
	664: "Seychelles",
	665: "St. Helena (UK)",
	666: "Somalia",
	667: "Sierra Leone",
	668: "Sao Tome & Principe",
	669: "Swaziland",
	670: "Chad",
	671: "Togolese Republic",
	672: "Tunisia",
	674: "Tanzania",
	675: "Uganda",
	676: "Congo",
	677: "Tanzania",
	678: "Zambia",
	679: "Zimbabwe",
	701: "Argentina",
	710: "Brazil",
	720: "Bolivia",
	725: "Chile",
	730: "Colombia",
	735: "Ecuador",
	740: "Falkland Islands (UK)",
	745: "Guiana (FR)",
	750: "Guyana",
	755: "Paraguay",
	760: "Peru",
	765: "Suriname",
	770: "Uruguay)",
}
