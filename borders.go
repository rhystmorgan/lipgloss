package lipgloss

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/clipperhouse/displaywidth"
	"github.com/muesli/termenv"
)

// Border contains a series of values which comprise the various parts of a
// border.
type Border struct {
	Top          string
	Bottom       string
	Left         string
	Right        string
	TopLeft      string
	TopRight     string
	BottomLeft   string
	BottomRight  string
	MiddleLeft   string
	MiddleRight  string
	Middle       string
	MiddleTop    string
	MiddleBottom string
}

// BorderSide represents a side of the block. It's used in selection, alignment,
// joining, placement and so on.
type BorderSide int

// BorderSide instances.
const (
	BorderTop BorderSide = iota
	BorderRight
	BorderBottom
	BorderLeft
)

// BorderHorizontalFunc is border function that sets horizontal border text
// at the configured position.
//
// It takes the width of the border and the Top/Bottom border string
// and determines the string for that position.
//
// Example:
//
//	reverseStyle := lipgloss.NewStyle().Reverse(true)
//	t := lipgloss.NewStyle().
//	    Border(lipgloss.NormalBorder()).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderTop,
//	        lipgloss.Center,
//	        func(w int, m string) string {
//	            return reverseStyle.Render(" BIG TITLE ")
//	        },
//	    )).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderBottom,
//	        lipgloss.Right,
//	        func(width int, middle string) string {
//	            return reverseStyle.Render(fmt.Sprintf(" %d/%d ", m.index + 1, m.count)) + middle
//	        },
//	    )).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderBottom,
//	        lipgloss.Left,
//	        func(width int, middle string) string {
//	            return middle + reverseStyle.Render(fmt.Sprintf("Status: %s", m.status))
//	        },
//	    ))
type BorderHorizontalFunc interface {
	func(width int, middle string) string
}

// BorderVerticalFunc is border function that sets vertical border text
// at the configured position.
//
// The first argument is the current row index, the second argument is
// the height of the border and the third is the Left/Right border string.
// It should return the border string for the given row.
//
// Example:
//
//	reverseStyle := lipgloss.NewStyle().Reverse(true)
//	t := lipgloss.NewStyle().
//	    Border(lipgloss.NormalBorder()).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderLeft,
//	        lipgloss.Top,
//	        func(row, height int, m string) string {
//	            if row % 2 == 1 {
//	                return "X"
//	            }
//	            return m
//	        },
//	    )).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderRight,
//	        lipgloss.Top,
//	        func(row, height int, middle string) string {
//	            if row == index {
//	                return "<"
//	            }
//	            return middle
//	        },
//	    ))
type BorderVerticalFunc interface {
	func(row, height int, middle string) string
}

// BorderDecoration is type used by Border to set text or decorate the border.
type BorderDecoration struct {
	side  BorderSide
	align Position
	st    interface{}
}

// BorderDecorator is constraint type for a string or function that is used
// to decorate a border.
type BorderDecorator interface {
	string | func() string | BorderHorizontalFunc | BorderVerticalFunc
}

// NewBorderDecoration is function that sets creates a decoration for the border.
//
// It takes the side of the border (Top|Bottom), the alignment (Left|Center|Right) of the
// decoration, and the decoration.
//
// the decoration can be any of
//   - string
//   - func() string
//   - func(width int, middle string) string  where width is the size of the border and middle
//     is the border string
//
// Example:
//
//	reverseStyle := lipgloss.NewStyle().Reverse(true)
//
//	t := lipgloss.NewStyle().
//	    Border(lipgloss.NormalBorder()).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderTop,
//	        lipgloss.Center,
//	        reverseStyle.Padding(0, 1).Render("BIG TITLE"),
//	    )).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderBottom,
//	        lipgloss.Right,
//	        func(width int, middle string) string {
//	            return reverseStyle.Render(fmt.Sprintf(" %d/%d ", m.index + 1, m.count)) + middle
//	        },
//	    )).
//	    BorderDecoration(lipgloss.NewBorderDecoration(
//	        lipgloss.BorderBottom,
//	        lipgloss.Left,
//	        reverseStyle.SetString(fmt.Sprintf("Status: %s", m.status)).String,
//	    ))
func NewBorderDecoration[S BorderDecorator](side BorderSide, align Position, st S) BorderDecoration {
	return BorderDecoration{
		align: align,
		side:  side,
		st:    st,
	}
}

// GetTopSize returns the width of the top border. If borders contain runes of
// varying widths, the widest rune is returned. If no border exists on the top
// edge, 0 is returned.
func (b Border) GetTopSize() int {
	return getBorderEdgeWidth(b.TopLeft, b.Top, b.TopRight)
}

// GetRightSize returns the width of the right border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the right edge, 0 is returned.
func (b Border) GetRightSize() int {
	return getBorderEdgeWidth(b.TopRight, b.Right, b.BottomRight)
}

// GetBottomSize returns the width of the bottom border. If borders contain
// runes of varying widths, the widest rune is returned. If no border exists on
// the bottom edge, 0 is returned.
func (b Border) GetBottomSize() int {
	return getBorderEdgeWidth(b.BottomLeft, b.Bottom, b.BottomRight)
}

// GetLeftSize returns the width of the left border. If borders contain runes
// of varying widths, the widest rune is returned. If no border exists on the
// left edge, 0 is returned.
func (b Border) GetLeftSize() int {
	return getBorderEdgeWidth(b.TopLeft, b.Left, b.BottomLeft)
}

func getBorderEdgeWidth(borderParts ...string) (maxWidth int) {
	for _, piece := range borderParts {
		maxWidth = max(maxWidth, maxRuneWidth(piece))
	}
	return maxWidth
}

var (
	noBorder = Border{}

	normalBorder = Border{
		Top:          "─",
		Bottom:       "─",
		Left:         "│",
		Right:        "│",
		TopLeft:      "┌",
		TopRight:     "┐",
		BottomLeft:   "└",
		BottomRight:  "┘",
		MiddleLeft:   "├",
		MiddleRight:  "┤",
		Middle:       "┼",
		MiddleTop:    "┬",
		MiddleBottom: "┴",
	}

	roundedBorder = Border{
		Top:          "─",
		Bottom:       "─",
		Left:         "│",
		Right:        "│",
		TopLeft:      "╭",
		TopRight:     "╮",
		BottomLeft:   "╰",
		BottomRight:  "╯",
		MiddleLeft:   "├",
		MiddleRight:  "┤",
		Middle:       "┼",
		MiddleTop:    "┬",
		MiddleBottom: "┴",
	}

	blockBorder = Border{
		Top:          "█",
		Bottom:       "█",
		Left:         "█",
		Right:        "█",
		TopLeft:      "█",
		TopRight:     "█",
		BottomLeft:   "█",
		BottomRight:  "█",
		MiddleLeft:   "█",
		MiddleRight:  "█",
		Middle:       "█",
		MiddleTop:    "█",
		MiddleBottom: "█",
	}

	outerHalfBlockBorder = Border{
		Top:         "▀",
		Bottom:      "▄",
		Left:        "▌",
		Right:       "▐",
		TopLeft:     "▛",
		TopRight:    "▜",
		BottomLeft:  "▙",
		BottomRight: "▟",
	}

	innerHalfBlockBorder = Border{
		Top:         "▄",
		Bottom:      "▀",
		Left:        "▐",
		Right:       "▌",
		TopLeft:     "▗",
		TopRight:    "▖",
		BottomLeft:  "▝",
		BottomRight: "▘",
	}

	thickBorder = Border{
		Top:          "━",
		Bottom:       "━",
		Left:         "┃",
		Right:        "┃",
		TopLeft:      "┏",
		TopRight:     "┓",
		BottomLeft:   "┗",
		BottomRight:  "┛",
		MiddleLeft:   "┣",
		MiddleRight:  "┫",
		Middle:       "╋",
		MiddleTop:    "┳",
		MiddleBottom: "┻",
	}

	doubleBorder = Border{
		Top:          "═",
		Bottom:       "═",
		Left:         "║",
		Right:        "║",
		TopLeft:      "╔",
		TopRight:     "╗",
		BottomLeft:   "╚",
		BottomRight:  "╝",
		MiddleLeft:   "╠",
		MiddleRight:  "╣",
		Middle:       "╬",
		MiddleTop:    "╦",
		MiddleBottom: "╩",
	}

	hiddenBorder = Border{
		Top:          " ",
		Bottom:       " ",
		Left:         " ",
		Right:        " ",
		TopLeft:      " ",
		TopRight:     " ",
		BottomLeft:   " ",
		BottomRight:  " ",
		MiddleLeft:   " ",
		MiddleRight:  " ",
		Middle:       " ",
		MiddleTop:    " ",
		MiddleBottom: " ",
	}

	markdownBorder = Border{
		Top:          "-",
		Bottom:       "-",
		Left:         "|",
		Right:        "|",
		TopLeft:      "|",
		TopRight:     "|",
		BottomLeft:   "|",
		BottomRight:  "|",
		MiddleLeft:   "|",
		MiddleRight:  "|",
		Middle:       "|",
		MiddleTop:    "|",
		MiddleBottom: "|",
	}

	asciiBorder = Border{
		Top:          "-",
		Bottom:       "-",
		Left:         "|",
		Right:        "|",
		TopLeft:      "+",
		TopRight:     "+",
		BottomLeft:   "+",
		BottomRight:  "+",
		MiddleLeft:   "+",
		MiddleRight:  "+",
		Middle:       "+",
		MiddleTop:    "+",
		MiddleBottom: "+",
	}
)

// NormalBorder returns a standard-type border with a normal weight and 90
// degree corners.
func NormalBorder() Border {
	return normalBorder
}

// RoundedBorder returns a border with rounded corners.
func RoundedBorder() Border {
	return roundedBorder
}

// BlockBorder returns a border that takes the whole block.
func BlockBorder() Border {
	return blockBorder
}

// OuterHalfBlockBorder returns a half-block border that sits outside the frame.
func OuterHalfBlockBorder() Border {
	return outerHalfBlockBorder
}

// InnerHalfBlockBorder returns a half-block border that sits inside the frame.
func InnerHalfBlockBorder() Border {
	return innerHalfBlockBorder
}

// ThickBorder returns a border that's thicker than the one returned by
// NormalBorder.
func ThickBorder() Border {
	return thickBorder
}

// DoubleBorder returns a border comprised of two thin strokes.
func DoubleBorder() Border {
	return doubleBorder
}

// HiddenBorder returns a border that renders as a series of single-cell
// spaces. It's useful for cases when you want to remove a standard border but
// maintain layout positioning. This said, you can still apply a background
// color to a hidden border.
func HiddenBorder() Border {
	return hiddenBorder
}

// MarkdownBorder return a table border in markdown style.
//
// Make sure to disable top and bottom border for the best result. This will
// ensure that the output is valid markdown.
//
//	table.New().Border(lipgloss.MarkdownBorder()).BorderTop(false).BorderBottom(false)
func MarkdownBorder() Border {
	return markdownBorder
}

// ASCIIBorder returns a table border with ASCII characters.
func ASCIIBorder() Border {
	return asciiBorder
}

func (s Style) applyBorder(str string) string {
	var (
		border    = s.getBorderStyle()
		hasTop    = s.getAsBool(borderTopKey, false)
		hasRight  = s.getAsBool(borderRightKey, false)
		hasBottom = s.getAsBool(borderBottomKey, false)
		hasLeft   = s.getAsBool(borderLeftKey, false)

		topFG    = s.getAsColor(borderTopForegroundKey)
		rightFG  = s.getAsColor(borderRightForegroundKey)
		bottomFG = s.getAsColor(borderBottomForegroundKey)
		leftFG   = s.getAsColor(borderLeftForegroundKey)

		topBG    = s.getAsColor(borderTopBackgroundKey)
		rightBG  = s.getAsColor(borderRightBackgroundKey)
		bottomBG = s.getAsColor(borderBottomBackgroundKey)
		leftBG   = s.getAsColor(borderLeftBackgroundKey)

		topFuncs    = s.borderTopFunc
		bottomFuncs = s.borderBottomFunc
		leftFuncs   = s.borderLeftFunc
		rightFuncs  = s.borderRightFunc
	)

	// If a border is set and no sides have been specifically turned on or off
	// render borders on all sides.
	if s.implicitBorders() {
		hasTop = true
		hasRight = true
		hasBottom = true
		hasLeft = true
	}

	// If no border is set or all borders are been disabled, abort.
	if border == noBorder || (!hasTop && !hasRight && !hasBottom && !hasLeft) {
		return str
	}

	lines, width := getLines(str)

	if hasLeft {
		if border.Left == "" {
			border.Left = " "
		}
		// width += maxRuneWidth(border.Left)
	}

	if hasRight && border.Right == "" {
		border.Right = " "
	}

	// If corners should be rendered but are set with the empty string, fill them
	// with a single space.
	if hasTop && hasLeft && border.TopLeft == "" {
		border.TopLeft = " "
	}
	if hasTop && hasRight && border.TopRight == "" {
		border.TopRight = " "
	}
	if hasBottom && hasLeft && border.BottomLeft == "" {
		border.BottomLeft = " "
	}
	if hasBottom && hasRight && border.BottomRight == "" {
		border.BottomRight = " "
	}

	// Figure out which corners we should actually be using based on which
	// sides are set to show.
	if hasTop {
		switch {
		case !hasLeft && !hasRight:
			border.TopLeft = ""
			border.TopRight = ""
		case !hasLeft:
			border.TopLeft = ""
		case !hasRight:
			border.TopRight = ""
		}
	}
	if hasBottom {
		switch {
		case !hasLeft && !hasRight:
			border.BottomLeft = ""
			border.BottomRight = ""
		case !hasLeft:
			border.BottomLeft = ""
		case !hasRight:
			border.BottomRight = ""
		}
	}

	// For now, limit corners to one rune.
	border.TopLeft = getFirstRuneAsString(border.TopLeft)
	border.TopRight = getFirstRuneAsString(border.TopRight)
	border.BottomRight = getFirstRuneAsString(border.BottomRight)
	border.BottomLeft = getFirstRuneAsString(border.BottomLeft)

	var out strings.Builder

	// Render top
	if hasTop {
		var top string

		if len(topFuncs) > 0 {
			top = renderAnnotatedHorizontalEdge(
				border.TopLeft,
				border.Top,
				border.TopRight,
				topFuncs,
				width,
				topFG,
				topBG,
				s.styleBorder,
			)
		} else {
			top = renderHorizontalEdge(
				border.TopLeft,
				border.Top,
				border.TopRight,
				width,
			)
		}
		// top = renderHorizontalEdge(border.TopLeft, border.Top, border.TopRight, width)
		top = s.styleBorder(top, topFG, topBG)
		out.WriteString(top)
		out.WriteRune('\n')
	}

	// leftRunes := []rune(border.Left)
	// leftIndex := 0

	leftBorder := make([]string, len(lines))
	rightBorder := make([]string, len(lines))

	if hasLeft {
		leftRunes := make([]rune, 0, len(lines))
		for len(leftRunes) < len(lines) {
			left := []rune(border.Left)
			leftRunes = append(leftRunes, left...)
		}
		leftRunes = leftRunes[:len(lines)]
		for i := range leftRunes {
			leftBorder[i] = string(leftRunes[i])
		}
		if len(leftFuncs) > 0 {
			leftBorder = renderVerticalEdge(
				leftBorder,
				border.Left,
				leftFuncs,
			)
		}
	}

	if hasRight {
		rightRunes := make([]rune, 0, len(lines))
		right := []rune(border.Right)
		for len(rightRunes) < len(lines) {
			rightRunes = append(rightRunes, right...)
		}
		rightRunes = rightRunes[:len(lines)]
		for i := range rightRunes {
			rightBorder[i] = string(rightRunes[i])
		}
		if len(rightFuncs) > 0 {
			rightBorder = renderVerticalEdge(
				rightBorder,
				border.Right,
				rightFuncs,
			)
		}
	}

	// rightRunes := []rune(border.Right)
	// rightIndex := 0

	// Render sides
	for i, l := range lines {
		if i > 0 {
			out.WriteRune('\n')
		}
		if hasLeft {
			out.WriteString(s.styleBorder(leftBorder[i], leftFG, leftBG))
		}
		out.WriteString(l)
		if hasRight {
			out.WriteString(s.styleBorder(rightBorder[i], rightFG, rightBG))
		}
	}

	// Render bottom
	if hasBottom {
		var bottom string
		if len(bottomFuncs) > 0 {
			bottom = renderAnnotatedHorizontalEdge(
				border.BottomLeft,
				border.Bottom,
				border.BottomRight,
				bottomFuncs,
				width,
				bottomFG,
				bottomBG,
				s.styleBorder,
			)
		} else {
			bottom = renderHorizontalEdge(border.BottomLeft, border.Bottom, border.BottomRight, width)
		}
		// bottom := renderHorizontalEdge(border.BottomLeft, border.Bottom, border.BottomRight, width)
		bottom = s.styleBorder(bottom, bottomFG, bottomBG)
		out.WriteRune('\n')
		out.WriteString(bottom)
	}

	return out.String()
}

// truncateWidths return the widths truncated to fit in the given
// length.
func truncateWidths(leftWidth, centerWidth, rightWidth, length int) (int, int, int) {
	leftWidth = min(leftWidth, length)
	centerWidth = min(centerWidth, length)
	rightWidth = min(rightWidth, length)

	if leftWidth == 0 && rightWidth == 0 {
		return leftWidth, centerWidth, rightWidth
	}

	if centerWidth == 0 {
		if leftWidth == 0 || rightWidth == 0 || leftWidth+rightWidth < length {
			return leftWidth, centerWidth, rightWidth
		}
		for leftWidth+rightWidth >= length {
			if leftWidth > rightWidth {
				leftWidth--
			} else {
				rightWidth--
			}
		}
		return leftWidth, centerWidth, rightWidth
	}

	for leftWidth >= length/2-(centerWidth+1)/2 {
		if leftWidth > centerWidth {
			leftWidth--
		} else {
			centerWidth--
		}
	}

	for rightWidth >= (length+1)/2-centerWidth/2 {
		if rightWidth > centerWidth {
			rightWidth--
		} else {
			centerWidth--
		}
	}

	return leftWidth, centerWidth, rightWidth
}

func renderVerticalEdge(edge []string, middle string, bFuncs []interface{}) []string {
	height := len(edge)

	var transformer func(int, int, string) string

	ts := make([]string, 3)
	ws := make([]int, 3)

	// get the decoration strings and truncate to fit within
	// the width.
	{
		for i, f := range bFuncs {
			if f == nil {
				continue
			}
			switch f := f.(type) {
			case string:
				ts[i] = f
			case func() string:
				ts[i] = f()
			case func(int, string) string:
				ts[i] = f(height, middle)
			case func(int, int, string) string:
				transformer = f
			}
			ws[i] = ansi.StringWidth(ts[i])
		}
		ws[0], ws[1], ws[2] = truncateWidths(ws[0], ws[1], ws[2], height)
		for i := range ts {
			ts[i] = ansi.Truncate(ts[i], ws[i], "")
		}
	}

	if ws[0] > 0 {
		copy(edge[0:], splitStyledString(ts[0]))
	}
	if ws[1] > 0 {
		copy(edge[(height-ws[1])/2:], splitStyledString(ts[1]))
	}
	if ws[2] > 0 {
		copy(edge[height-ws[2]:], splitStyledString(ts[2]))
	}

	if transformer != nil {
		// transform
		for i := range edge {
			w := ansi.StringWidth(edge[i])
			edge[i] = transformer(i, height, edge[i])
			edge[i] = ansi.Truncate(edge[i], w, "")
		}
	}

	return edge
}

func renderAnnotatedHorizontalEdge(
	left, middle, right string,
	bFuncs []interface{},
	width int,
	borderFG, borderBG TerminalColor,
	styleBorderFunc func(string, TerminalColor, TerminalColor) string,
) string {
	if middle == "" {
		middle = " "
	}

	//Extract decoration content
	decorations := make([]string, 3)
	for i, f := range bFuncs {
		if f == nil {
			continue
		}
		switch f := f.(type) {
		case string:
			decorations[i] = f
		case func() string:
			decorations[i] = f()
		case func(int, string) string:
			decorations[i] = f(width, middle)
		}
	}

	leftDecor := decorations[0]
	centerDecor := decorations[1]
	rightDecor := decorations[2]

	// Calculate visual widths
	leftWidth := ansi.StringWidth(leftDecor)
	centerWidth := ansi.StringWidth(centerDecor)
	rightWidth := ansi.StringWidth(rightDecor)

	// Calculate positions
	// Position 0: left corner
	// Position width-1: right corner
	// Left decoration: positions [1, leftWidth]
	// Center decoration: positions [centerStart, centerStart + centerWidth]
	// Right decoration: positions [width-1-rightWidth, width-1)
	centerStart := (width - centerWidth) / 2
	rightStart := width - 1 - rightWidth
	middleRune := []rune(middle)[0]

	var result strings.Builder

	// Position 0: left corner
	result.WriteString(styleBorderFunc(left, borderFG, borderBG))

	// Left decoration
	if leftWidth > 0 {
		result.WriteString(leftDecor)
	}

	// Middle section from after left decoration to center decoration
	for i := leftWidth; i < centerStart; i++ {
		result.WriteString(styleBorderFunc(string(middleRune), borderFG, borderBG))
	}

	// Center decoration (render as-is, don't modify styling)
	if centerWidth > 0 {
		result.WriteString(centerDecor)
	}

	// Middle section from after center decoration to right decoration
	for i := centerStart + centerWidth; i <= rightStart; i++ {
		result.WriteString(styleBorderFunc(string(middleRune), borderFG, borderBG))
	}

	// Right decoration
	if rightWidth > 0 {
		result.WriteString(rightDecor)
	}

	// Position width-1: right corner
	result.WriteString(styleBorderFunc(right, borderFG, borderBG))
	return result.String()
}

// Render the horizontal (top or bottom) portion of a border.
func renderHorizontalEdge(left, middle, right string, width int) string {
	if middle == "" {
		middle = " "
	}

	runes := []rune(middle)
	j := 0

	out := strings.Builder{}
	out.WriteString(left)
	for i := 0; i < width; {
		out.WriteRune(runes[j])
		j++
		if j >= len(runes) {
			j = 0
		}
		i += ansi.StringWidth(string(runes[j]))
	}
	out.WriteString(right)

	return out.String()
}

// Apply foreground and background styling to a border.
func (s Style) styleBorder(border string, fg, bg TerminalColor) string {
	if fg == noColor && bg == noColor {
		return border
	}

	style := termenv.Style{}

	if fg != noColor {
		style = style.Foreground(fg.color(s.r))
	}
	if bg != noColor {
		style = style.Background(bg.color(s.r))
	}

	return style.Styled(border)
}

func maxRuneWidth(str string) int {
	switch len(str) {
	case 0:
		return 0
	case 1:
		return displaywidth.String(str)
	}

	var width int

	g := displaywidth.StringGraphemes(str)
	for g.Next() {
		width = max(width, g.Width())
	}
	return width
}

func getFirstRuneAsString(str string) string {
	if str == "" {
		return str
	}
	_, size := utf8.DecodeRuneInString(str)
	return str[:size]
}

// Alternative: Split styled string character by character, preserving ANSI codes
func splitStyledString(s string) []string {
	if s == "" {
		return []string{}
	}

	x := ansi.Strip(s)
	if x == s {
		// No styles, just split into characters
		return strings.Split(s, "")
	}

	// For styled strings, we need to wrap each character with its ansi codes
	// extract the prefix
	prefix := ""
	if i := strings.Index(s, x); i > 0 {
		prefix = s[:i]
		s = s[i:]
	}

	// Extract the suffix
	suffix := ""
	if i := strings.Index(s, x); i > 0 {
		suffix = s[i:]
		s = s[i:]
	}

	// each character gets: prefix b char - suffix
	result := make([]string, len(x))
	for i, r := range x {
		result[i] = prefix + string(r) + suffix
	}

	return result
}
