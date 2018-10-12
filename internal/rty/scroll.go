package rty

import (
	"strings"
)

type StatefulComponent interface {
	RenderStateful(w Writer, prevState interface{}, width, height int) (state interface{}, err error)
}

type TextScrollLayout struct {
	name string
	cs   []Component
}

func NewTextScrollLayout(name string) *TextScrollLayout {
	return &TextScrollLayout{name: name}
}

func (l *TextScrollLayout) Add(c Component) {
	l.cs = append(l.cs, c)
}

func (l *TextScrollLayout) Size(width int, height int) (int, int) {
	return width, height
}

type TextScrollState struct {
	width  int
	height int

	canvasIdx     int
	lineIdx       int
	canvasLengths []int
}

func (l *TextScrollLayout) Render(w Writer, width, height int) error {
	w.RenderStateful(l, l.name)
	return nil
}

func (l *TextScrollLayout) RenderStateful(w Writer, prevState interface{}, width, height int) (state interface{}, err error) {
	prev, ok := prevState.(*TextScrollState)
	if !ok {
		prev = &TextScrollState{}
	}
	next := &TextScrollState{
		width:  width,
		height: height,
	}

	if len(l.cs) == 0 {
		return next, nil
	}

	next.canvasLengths = make([]int, len(l.cs))
	canvases := make([]Canvas, len(l.cs))

	for i, c := range l.cs {
		childCanvas := w.RenderChildInTemp(c)
		canvases[i] = childCanvas
		_, childHeight := childCanvas.Size()
		next.canvasLengths[i] = childHeight
	}

	l.adjustCursor(prev, next, canvases)

	y := 0
	canvases = canvases[next.canvasIdx:]

	if next.lineIdx != 0 {
		firstCanvas := canvases[0]
		canvases = canvases[1:]
		_, firstHeight := firstCanvas.Size()
		numLines := firstHeight - prev.lineIdx
		if numLines > height {
			numLines = height
		}

		w.Divide(0, 0, width, numLines).Embed(firstCanvas, next.lineIdx, numLines)
		y += numLines
	}

	for _, canvas := range canvases {
		_, canvasHeight := canvas.Size()
		numLines := canvasHeight
		if numLines > height-y {
			numLines = height - y
		}
		w.Divide(0, y, width, numLines).Embed(canvas, 0, numLines)
		y += numLines
	}

	return next, nil
}

func (l *TextScrollLayout) adjustCursor(prev *TextScrollState, next *TextScrollState, canvases []Canvas) {
	if prev.canvasIdx >= len(canvases) {
		return
	}

	next.canvasIdx = prev.canvasIdx
	_, canvasHeight := canvases[next.canvasIdx].Size()
	if prev.lineIdx >= canvasHeight {
		return
	}
	next.lineIdx = prev.lineIdx
}

type TextScrollController struct {
	state *TextScrollState
}

func NewTextScrollController(state *TextScrollState) *TextScrollController {
	return &TextScrollController{state: state}
}

func (s *TextScrollController) Up() {
	st := s.state
	if st.lineIdx != 0 {
		st.lineIdx--
		return
	}

	if st.canvasIdx == 0 {
		return
	}
	st.canvasIdx--
	st.lineIdx = st.canvasLengths[st.canvasIdx] - 1
}

func (s *TextScrollController) Down() {
	st := s.state
	canvasLength := st.canvasLengths[st.canvasIdx]
	if st.lineIdx < canvasLength-1 {
		// we can just go down in this canvas
		st.lineIdx++
		return
	}
	if st.canvasIdx == len(st.canvasLengths)-1 {
		// we're at the end of the last canvas
		return
	}
	st.canvasIdx++
	st.lineIdx = 0
}

func NewScrollingWrappingTextArea(name string, text string) Component {
	l := NewTextScrollLayout(name)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		l.Add(NewWrappingTextLine(line))
	}

	return l
}