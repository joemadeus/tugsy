package main

import "github.com/veandco/go-sdl2/sdl"

type Render func(history *ShipHistory)

type RenderStyle interface {
	Render(view *View) Render
}

type NullRenderStyle struct{}

func (style *NullRenderStyle) Render(view *View) Render {
	return func(history *ShipHistory) {}
}

type MarkCurrentPositionSimple struct{}

func (style *MarkCurrentPositionSimple) Render(view *View) Render {
	currentPositionR := uint8(128)
	currentPositionG := uint8(128)
	currentPositionB := uint8(0)
	currentPositionSize := int32(10)
	return func(history *ShipHistory) {
		currentPosition := history.positions[len(history.positions)-1]
		baseMapPosition := view.getBaseMapPosition(currentPosition.GetPositionReport())

		view.screenRenderer.SetDrawColor(currentPositionR, currentPositionG, currentPositionB, sdl.ALPHA_OPAQUE)
		err := view.screenRenderer.FillRect(
			&sdl.Rect{
				X: int32(baseMapPosition.X+0.5) - (currentPositionSize / 2),
				Y: int32(baseMapPosition.Y+0.5) - (currentPositionSize / 2),
				W: currentPositionSize,
				H: currentPositionSize})
		if err != nil {
			logger.Warn("rendering current position", "error", err)
		}
	}
}

type MarkPathSimple struct{}

func (style *MarkPathSimple) Render(view *View) Render {
	defaultTrackLinesR := uint8(192)
	defaultTrackLinesG := uint8(192)
	defaultTrackLinesB := uint8(0)

	defaultTrackPointsR := uint8(128)
	defaultTrackPointsG := uint8(128)
	defaultTrackPointsB := uint8(0)

	trackPointsSize := int32(4)

	return func(history *ShipHistory) {
		sdlPoints := make([]sdl.Point, len(history.positions), len(history.positions))
		sdlRects := make([]sdl.Rect, len(history.positions), len(history.positions))

		for i, position := range history.positions {
			baseMapPosition := view.getBaseMapPosition(position.GetPositionReport())
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

		view.screenRenderer.SetDrawColor(defaultTrackLinesR, defaultTrackLinesG, defaultTrackLinesB, sdl.ALPHA_OPAQUE)
		err := view.screenRenderer.DrawLines(sdlPoints)
		if err != nil {
			logger.Warn("rendering track lines", "error", err)
		}

		view.screenRenderer.SetDrawColor(defaultTrackPointsR, defaultTrackPointsG, defaultTrackPointsB, sdl.ALPHA_OPAQUE)
		err = view.screenRenderer.DrawRects(sdlRects)
		if err != nil {
			logger.Warn("rendering track points", "error", err)
		}
	}
}
