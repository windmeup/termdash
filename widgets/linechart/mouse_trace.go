package linechart

import (
	"fmt"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/linechart/internal/axes"
	"image"
)

type MouseTrace struct {
	*LineChart

	mousePoint *image.Point
}

func NewMouseTrace(opts ...Option) (*MouseTrace, error) {
	lc, err := New(opts...)
	if err != nil {
		return nil, err
	}
	return &MouseTrace{
		LineChart:  lc,
		mousePoint: &image.Point{X: 10, Y: 10}, // TODO test
	}, nil
}

func (lc *MouseTrace) Draw(cvs *canvas.Canvas, meta *widgetapi.Meta) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	needAr, err := area.FromSize(lc.minSize())
	if err != nil {
		return err
	}
	if !needAr.In(cvs.Area()) {
		return draw.ResizeNeeded(cvs)
	}

	xd, yd, err := lc.axesDetails(cvs)
	if err != nil {
		return err
	}

	adjXD, err := lc.drawSeries(cvs, xd, yd)
	if err != nil {
		return err
	}

	if err := lc.drawMouse(cvs, adjXD, yd); err != nil {
		return err
	}

	if err := lc.drawAxes(cvs, adjXD, yd); err != nil {
		return err
	}
	return nil
}

func (lc *MouseTrace) drawMouse(cvs *canvas.Canvas, xd *axes.XDetails, yd *axes.YDetails) error {
	if lc.mousePoint == nil {
		return nil
	}
	if err := draw.HVLines(cvs, []draw.HVLine{
		{Start: image.Point{X: lc.mousePoint.X, Y: yd.Start.Y}, End: image.Point{X: lc.mousePoint.X, Y: yd.End.Y}},
	}, draw.HVLineCellOpts(cell.FgColor(cell.ColorGray))); err != nil {
		return fmt.Errorf("failed to draw mouse: %w", err)
	}
	return nil
}

func (lc *MouseTrace) Mouse(m *terminalapi.Mouse, meta *widgetapi.EventMeta) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.zoom == nil {
		return nil
	}
	if err := lc.zoom.Mouse(m); err != nil {
		return err
	}
	if !m.Position.In(lc.zoom.GraphArea()) {
		lc.mousePoint = nil
		return nil
	}
	if lc.mousePoint == nil {
		lc.mousePoint = &image.Point{
			X: m.Position.X,
		}
	} else {
		lc.mousePoint.X = m.Position.X
	}
	return nil
}
