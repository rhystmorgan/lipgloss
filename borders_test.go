package lipgloss

import (
	"testing"

	"github.com/rivo/uniseg"
)

func TestStyle_GetBorderSizes(t *testing.T) {
	tests := []struct {
		name  string
		style Style
		wantX int
		wantY int
	}{
		{
			name:  "Default style",
			style: NewStyle(),
			wantX: 0,
			wantY: 0,
		},
		{
			name:  "Border(NormalBorder())",
			style: NewStyle().Border(NormalBorder()),
			wantX: 2,
			wantY: 2,
		},
		{
			name:  "Border(NormalBorder(), true)",
			style: NewStyle().Border(NormalBorder(), true),
			wantX: 2,
			wantY: 2,
		},
		{
			name:  "Border(NormalBorder(), true, false)",
			style: NewStyle().Border(NormalBorder(), true, false),
			wantX: 0,
			wantY: 2,
		},
		{
			name:  "Border(NormalBorder(), true, true, false)",
			style: NewStyle().Border(NormalBorder(), true, true, false),
			wantX: 2,
			wantY: 1,
		},
		{
			name:  "Border(NormalBorder(), true, true, false, false)",
			style: NewStyle().Border(NormalBorder(), true, true, false, false),
			wantX: 1,
			wantY: 1,
		},
		{
			name:  "BorderTop(true).BorderStyle(NormalBorder())",
			style: NewStyle().BorderTop(true).BorderStyle(NormalBorder()),
			wantX: 0,
			wantY: 1,
		},
		{
			name:  "BorderStyle(NormalBorder())",
			style: NewStyle().BorderStyle(NormalBorder()),
			wantX: 2,
			wantY: 2,
		},
		{
			name:  "Custom BorderStyle",
			style: NewStyle().BorderStyle(Border{Left: "123456789"}),
			wantX: 1, // left and right borders are laid out vertically, one rune per row
			wantY: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX := tt.style.GetHorizontalBorderSize()
			if gotX != tt.wantX {
				t.Errorf("Style.GetHorizontalBorderSize() got %d, want %d", gotX, tt.wantX)
			}

			gotY := tt.style.GetVerticalBorderSize()
			if gotY != tt.wantY {
				t.Errorf("Style.GetVerticalBorderSize() got %d, want %d", gotY, tt.wantY)
			}

			gotX = tt.style.GetHorizontalFrameSize()
			if gotX != tt.wantX {
				t.Errorf("Style.GetHorizontalFrameSize() got %d, want %d", gotX, tt.wantX)
			}

			gotY = tt.style.GetVerticalFrameSize()
			if gotY != tt.wantY {
				t.Errorf("Style.GetVerticalFrameSize() got %d, want %d", gotY, tt.wantY)
			}

			gotX, gotY = tt.style.GetFrameSize()
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Errorf("Style.GetFrameSize() got (%d, %d), want (%d, %d)", gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

// Old implementation using rune slice conversion
func getFirstRuneAsStringOld(str string) string {
	if str == "" {
		return str
	}
	r := []rune(str)
	return string(r[0])
}

func TestGetFirstRuneAsString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty", "", ""},
		{"SingleASCII", "A", "A"},
		{"SingleUnicode", "世", "世"},
		{"ASCIIString", "Hello", "H"},
		{"UnicodeString", "你好世界", "你"},
		{"MixedASCIIFirst", "Hello世界", "H"},
		{"MixedUnicodeFirst", "世界Hello", "世"},
		{"Emoji", "😀Happy", "😀"},
		{"MultiByteFirst", "ñoño", "ñ"},
		{"LongString", "The quick brown fox jumps over the lazy dog", "T"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFirstRuneAsString(tt.input)
			if got != tt.want {
				t.Errorf("getFirstRuneAsString(%q) = %q, want %q", tt.input, got, tt.want)
			}

			// Verify new implementation matches old implementation
			old := getFirstRuneAsStringOld(tt.input)
			if got != old {
				t.Errorf("getFirstRuneAsString(%q) = %q, but old implementation returns %q", tt.input, got, old)
			}
		})
	}
}

func BenchmarkGetFirstRuneAsString(b *testing.B) {
	testCases := []struct {
		name string
		str  string
	}{
		{"ASCII", "Hello, World!"},
		{"Unicode", "你好世界"},
		{"Single", "A"},
		{"Empty", ""},
	}

	b.Run("Old", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = getFirstRuneAsStringOld(tc.str)
				}
			})
		}
	})

	b.Run("New", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_ = getFirstRuneAsString(tc.str)
				}
			})
		}
	})
}

func BenchmarkMaxRuneWidth(b *testing.B) {
	testCases := []struct {
		name string
		str  string
	}{
		{"Blank", " "},
		{"ASCII", "+"},
		{"Markdown", "|"},
		{"Normal", "├"},
		{"Rounded", "╭"},
		{"Block", "█"},
		{"Emoji", "😀"},
	}
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.Run("Before", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = maxRuneWidthOld(tc.str)
				}
			})
			b.Run("After", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = maxRuneWidth(tc.str)
				}
			})
		})
	}
}

func maxRuneWidthOld(str string) int {
	var width int

	state := -1
	for len(str) > 0 {
		var w int
		_, str, w, state = uniseg.FirstGraphemeClusterInString(str, state)
		if w > width {
			width = w
		}
	}

	return width
}

func TestBorderFunc(t *testing.T) {

	tt := []struct {
		name     string
		text     string
		style    Style
		expected string
	}{
		{
			name: "trunc all string",
			text: "",
			style: NewStyle().
				Width(16).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(BorderTop, Left, "LeftLeftLeftLeft")).
				BorderDecoration(NewBorderDecoration(BorderTop, Center, "CenterCenterCenter")).
				BorderDecoration(NewBorderDecoration(BorderTop, Right, "RightRightRightRight")),
			expected: `┌LeftL─Cent─Right┐
│                │
└────────────────┘`,
		},
		{
			name: "top left title string",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(BorderTop, Left, "TITLE")),
			expected: `┌TITLE─────┐
│          │
└──────────┘`,
		},
		{
			name: "top left title stringer",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(BorderTop, Left, NewStyle().SetString("TITLE").String)),
			expected: `┌TITLE─────┐
│          │
└──────────┘`,
		},
		{
			name: "top left very long title stringer",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(BorderTop, Left, NewStyle().SetString("TitleTitleTitle").String)),
			expected: `┌TitleTitle┐
│          │
└──────────┘`,
		},
		{
			name: "top left title",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderTop,
					Left,
					func(width int, middle string) string {
						return "TITLE"
					},
				)),
			expected: `┌TITLE─────┐
│          │
└──────────┘`,
		},
		{
			name: "top center title",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderTop,
					Center,
					func(width int, middle string) string {
						return "TITLE"
					},
				)),
			expected: `┌──TITLE───┐
│          │
└──────────┘`,
		},
		{
			name: "top center title even",
			text: "",
			style: NewStyle().
				Width(11).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderTop,
					Center,
					func(width int, middle string) string {
						return "TITLE"
					},
				)),
			expected: `┌───TITLE───┐
│           │
└───────────┘`,
		},
		{
			name: "top right title",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderTop,
					Right,
					func(width int, middle string) string {
						return "TITLE"
					},
				)),
			expected: `┌─────TITLE┐
│          │
└──────────┘`,
		},
		{
			name: "bottom left title",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderBottom,
					Left,
					func(width int, middle string) string {
						return "STATUS"
					},
				)),
			expected: `┌──────────┐
│          │
└STATUS────┘`,
		},
		{
			name: "bottom center title",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderBottom,
					Center,
					func(width int, middle string) string {
						return "STATUS"
					},
				)),
			expected: `┌──────────┐
│          │
└──STATUS──┘`,
		},
		{
			name: "bottom center title odd",
			text: "",
			style: NewStyle().
				Width(11).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderBottom,
					Center,
					func(width int, middle string) string {
						return "STATUS"
					},
				)),
			expected: `┌───────────┐
│           │
└──STATUS───┘`,
		},
		{
			name: "bottom right title",
			text: "",
			style: NewStyle().
				Width(10).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderBottom,
					Right,
					func(width int, middle string) string {
						return "STATUS"
					},
				)),
			expected: `┌──────────┐
│          │
└────STATUS┘`,
		},
		{
			name: "bottom right padded title",
			text: "",
			style: NewStyle().
				Width(12).
				Border(NormalBorder()).
				BorderDecoration(NewBorderDecoration(
					BorderBottom,
					Right,
					func(width int, middle string) string {
						return "│STATUS│" + middle
					},
				)),
			expected: `┌────────────┐
│            │
└───│STATUS│─┘`,
		},
	}

	for i, tc := range tt {
		res := tc.style.Render(tc.text)
		if res != tc.expected {
			t.Errorf("Test %d, expected:\n\n`%s`\n`%s`\n\nActual output:\n\n`%s`\n`%s`\n\n",
				i, tc.expected, formatEscapes(tc.expected),
				res, formatEscapes(res))
		}
	}

}

func TestBorders(t *testing.T) {
	tt := []struct {
		name     string
		text     string
		style    Style
		expected string
	}{
		{
			name:  "border with width",
			text:  "",
			style: NewStyle().Width(10).Border(NormalBorder()),
			expected: `┌──────────┐
│          │
└──────────┘`,
		},
		{
			name:  "top center title",
			text:  "HELLO",
			style: NewStyle().Border(NormalBorder()),
			expected: `┌─────┐
│HELLO│
└─────┘`,
		},
	}

	for i, tc := range tt {
		res := tc.style.Render(tc.text)
		if res != tc.expected {
			t.Errorf("Test %d, expected:\n\n`%s`\n`%s`\n\nActual output:\n\n`%s`\n`%s`\n\n",
				i, tc.expected, formatEscapes(tc.expected),
				res, formatEscapes(res))
		}
	}

}

func TestTruncateWidths(t *testing.T) {

	tt := []struct {
		name     string
		widths   [3]int
		width    int
		expected [3]int
	}{
		{
			name:     "lll-cc-rrr",
			widths:   [3]int{10, 10, 10},
			width:    10,
			expected: [3]int{3, 2, 3},
		},
		{
			name:     "lll-ccc-rrr",
			widths:   [3]int{10, 10, 10},
			width:    12,
			expected: [3]int{3, 3, 4},
		},
		{
			name:     "lllll-rrrr",
			widths:   [3]int{10, 0, 10},
			width:    10,
			expected: [3]int{5, 0, 4},
		},
		{
			name:     "lllllll-rr",
			widths:   [3]int{10, 0, 2},
			width:    10,
			expected: [3]int{7, 0, 2},
		},
		{
			name:     "ll-rrrrrrr",
			widths:   [3]int{2, 0, 20},
			width:    10,
			expected: [3]int{2, 0, 7},
		},
		{
			name:     "lll-cc----",
			widths:   [3]int{10, 10, 0},
			width:    10,
			expected: [3]int{3, 2, 0},
		},
		{
			name:     "----cc-rrr",
			widths:   [3]int{0, 10, 10},
			width:    10,
			expected: [3]int{0, 3, 3},
		},
	}

	for i, tc := range tt {
		var result [3]int

		result[0], result[1], result[2] = truncateWidths(tc.widths[0], tc.widths[1], tc.widths[2], tc.width)
		if result != tc.expected {
			t.Errorf("Test %d, expected:`%v`Actual output:`%v`", i, tc.expected, result)
		}
	}

}

func TestSplitStyledString(t *testing.T) {

	tt := []struct {
		input    string
		expected []string
	}{
		{
			input:    "abc",
			expected: []string{"a", "b", "c"},
		},
		{
			input:    "\x1b[41mabc\x1b[0m",
			expected: []string{"\x1b[41ma\x1b[0m", "\x1b[41mb\x1b[0m", "\x1b[41mc\x1b[0m"},
		},
		{
			input:    "VERTICAL",
			expected: []string{"V", "E", "R", "T", "I", "C", "A", "L"},
		},
	}

	for i, tc := range tt {
		got := splitStyledString(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("Test %d expected:`%v`Actual output:`%v`", i, tc.expected, got)
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("Item %d, expected:`%q`Actual output:`%q`", i, tc.expected[i], got[i])
			}
		}
	}
}
