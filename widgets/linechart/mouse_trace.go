package linechart

import (
	"fmt"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/mouse"
	"github.com/mum4k/termdash/private/area"
	"github.com/mum4k/termdash/private/canvas"
	"github.com/mum4k/termdash/private/draw"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgetapi"
	"github.com/mum4k/termdash/widgets/linechart/internal/axes"
	"image"
	"math"
	"sort"
	"strings"
)

type MouseTrace struct {
	*LineChart

	mousePoint *image.Point
	mouseValue string

	mouseIdx int
}

func NewMouseTrace(opts ...Option) (*MouseTrace, error) {
	lc, err := New(opts...)
	if err != nil {
		return nil, err
	}
	return &MouseTrace{
		LineChart: lc,
	}, nil
}

func (lc *MouseTrace) Draw(cvs *canvas.Canvas, _ *widgetapi.Meta) error {
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

	if err := lc.drawMouse(cvs, yd); err != nil {
		return err
	}

	if err := lc.drawAxes(cvs, adjXD, yd); err != nil {
		return err
	}
	if err := lc.drawMouseValue(cvs, yd); err != nil {
		return err
	}
	return nil
}

func (lc *MouseTrace) drawMouse(cvs *canvas.Canvas, yd *axes.YDetails) error {
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

func (lc *MouseTrace) drawMouseValue(cvs *canvas.Canvas, yd *axes.YDetails) error {
	pos := yd.Start.Add(image.Point{X: 2, Y: 0})
	if err := draw.Text(cvs, lc.mouseValue, pos, draw.TextCellOpts(cell.FgColor(cell.ColorWhite))); err != nil {
		return err
	}
	return nil
}

func (lc *MouseTrace) Mouse(m *terminalapi.Mouse, _ *widgetapi.EventMeta) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.zoom == nil {
		return nil
	}
	if err := lc.zoom.Mouse(m); err != nil {
		return err
	}
	graphArea := lc.zoom.GraphArea()
	if !m.Position.In(graphArea) {
		lc.mousePoint = nil
		lc.mouseValue = ""
		return nil
	}
	xdZoomed := lc.zoom.Zoom()
	v, err := xdZoomed.Scale.PixelToValue(m.Position.X - graphArea.Min.X)
	if err != nil {
		return err
	}
	idx := int(math.Round(v*2)) - int(xdZoomed.Scale.Min.Rounded) // tricky
	if idx < 0 {
		idx = 0
	}
	var label string
	type sV struct {
		name  string
		value float64
	}
	var mVs []sV
	for name, values := range lc.series {
		if idx < len(values.values) {
			if label == "" {
				if values.xLabelsSet {
					if l := values.xLabels[idx]; l != "" {
						label = l
					}
				}
				if m.Button == mouse.ButtonLeft {
					lc.mouseIdx = idx
				}
			}
			mVs = append(mVs, sV{name: name, value: values.values[idx]})
		}
	}
	if mVs == nil {
		lc.mouseValue = ""
	} else {
		sort.Slice(mVs, func(i, j int) bool {
			return strings.Compare(mVs[i].name, mVs[j].name) < 0
		})
		var mv string
		for _, v := range mVs {
			mv += fmt.Sprintf(" %s: %.2f", v.name, v.value) // TODO format option
		}
		lc.mouseValue = fmt.Sprintf("%s:%s", label, mv)
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

func (lc *MouseTrace) MouseIndex() int {
	return lc.mouseIdx
}
