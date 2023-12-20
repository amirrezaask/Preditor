package preditor

import (
	"bytes"
	"fmt"
	"github.com/amirrezaask/preditor/byteutils"
	"golang.design/x/clipboard"
	"os"
	"path"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type ScoredItem[T any] struct {
	Item  T
	Score int
}

func getClipboardContent() []byte {
	return clipboard.Read(clipboard.FmtText)
}

func writeToClipboard(bs []byte) {
	clipboard.Write(clipboard.FmtText, bytes.Clone(bs))
}

type List[T any] struct {
	BaseDrawable
	cfg                     *Config
	parent                  *Context
	keymaps                 []Keymap
	VisibleStart            int
	Items                   []T
	Selection               int
	maxHeight               int32
	maxWidth                int32
	UserInput               []byte
	ZeroLocation            rl.Vector2
	Idx                     int
	LastInput               string
	LastInputWeRanUpdateFor string
	UpdateList              func(list *List[T], input string)
	OpenSelection           func(ctx *Context, t T) error
	ItemRepr                func(item T) string
}

func (l *List[T]) SetNewUserInput(bs []byte) {
	l.LastInput = string(l.UserInput)
	l.UserInput = bs
	l.Idx += len(l.UserInput)

	if l.Idx >= len(l.UserInput) {
		l.Idx = len(l.UserInput)
	} else if l.Idx < 0 {
		l.Idx = 0
	}

}
func (l *List[T]) InsertCharAtBuffer(char byte) error {
	l.SetNewUserInput(append(l.UserInput, char))
	return nil
}

func (l *List[T]) CursorRight(n int) error {
	if l.Idx >= len(l.UserInput) {
		return nil
	}

	l.Idx += n

	return nil
}

func (l *List[T]) Paste() error {
	content := getClipboardContent()
	l.UserInput = append(l.UserInput[:l.Idx], append(content, l.UserInput[l.Idx+1:]...)...)

	return nil
}

func (l *List[T]) KillLine() error {
	l.SetNewUserInput(l.UserInput[:l.Idx])
	return nil
}

func (l *List[T]) Copy() error {
	writeToClipboard(l.UserInput)

	return nil
}

func (l *List[T]) BeginningOfTheLine() error {
	l.Idx = 0
	return nil
}

func (l *List[T]) EndOfTheLine() error {
	l.Idx = len(l.UserInput)
	return nil
}

func (l *List[T]) NextWordStart() error {
	if idx := byteutils.NextWordInBuffer(l.UserInput, l.Idx); idx != -1 {
		l.Idx = idx
	}

	return nil
}

func (l *List[T]) CursorLeft(n int) error {

	if l.Idx <= 0 {
		return nil
	}

	l.Idx -= n

	return nil
}

func (l *List[T]) PreviousWord() error {
	if idx := byteutils.PreviousWordInBuffer(l.UserInput, l.Idx); idx != -1 {
		l.Idx = idx
	}

	return nil
}

func (l *List[T]) DeleteCharBackward() error {
	if l.Idx <= 0 {
		return nil
	}
	if len(l.UserInput) <= l.Idx {
		l.SetNewUserInput(l.UserInput[:l.Idx-1])
	} else {
		l.SetNewUserInput(append(l.UserInput[:l.Idx-1], l.UserInput[l.Idx:]...))
	}
	return nil
}

func (l *List[T]) DeleteWordBackward() error {
	previousWordEndIdx := byteutils.PreviousWordInBuffer(l.UserInput, l.Idx)
	if len(l.UserInput) > l.Idx+1 {
		l.SetNewUserInput(append(l.UserInput[:previousWordEndIdx+1], l.UserInput[l.Idx+1:]...))
	} else {
		l.SetNewUserInput(l.UserInput[:previousWordEndIdx+1])
	}
	return nil
}
func (l *List[T]) DeleteWordForward() error {
	nextWordStartIdx := byteutils.NextWordInBuffer(l.UserInput, l.Idx)
	if len(l.UserInput) > nextWordStartIdx+1 {
		l.SetNewUserInput(append(l.UserInput[:l.Idx+1], l.UserInput[nextWordStartIdx+1:]...))
	} else {
		l.SetNewUserInput(l.UserInput[:l.Idx])
	}

	return nil
}
func (l *List[T]) DeleteCharForward() error {
	if l.Idx < 0 {
		return nil
	}
	l.SetNewUserInput(append(l.UserInput[:l.Idx], l.UserInput[l.Idx+1:]...))
	return nil
}

func (l *List[T]) NextItem() {
	l.Selection++
	if l.Selection >= len(l.Items) {
		l.Selection = len(l.Items) - 1
	}

}

func (l *List[T]) PrevItem() {
	l.Selection--
	if l.Selection < 0 {
		l.Selection = 0
	}

	if l.Selection < l.VisibleStart {
		l.VisibleStart--
		if l.VisibleStart < 0 {
			l.VisibleStart = 0
		}
	}

}
func (l *List[T]) Scroll(n int) {
	l.VisibleStart += n

	if l.VisibleStart < 0 {
		l.VisibleStart = 0
	}

}
func (l *List[T]) VisibleView(maxLine int) []T {
	if l.Selection < l.VisibleStart {
		l.VisibleStart -= maxLine / 3
		if l.VisibleStart < 0 {
			l.VisibleStart = 0
		}
	}

	if l.Selection >= l.VisibleStart+maxLine {
		l.VisibleStart += maxLine / 3
		if l.VisibleStart >= len(l.Items) {
			l.VisibleStart = len(l.Items)
		}
	}

	if len(l.Items) > l.VisibleStart+maxLine {
		return l.Items[l.VisibleStart : l.VisibleStart+maxLine]
	} else {
		return l.Items[l.VisibleStart:len(l.Items)]
	}
}

func (l List[T]) Keymaps() []Keymap {
	return l.keymaps
}

func (l List[T]) String() string {
	return fmt.Sprintf("List: %T", *new(T))
}

func NewList[T any](
	parent *Context,
	cfg *Config,
	updateList func(list *List[T], input string),
	openSelection func(preditor *Context, t T) error,
	repr func(t T) string,
	initialList func() []T,
) *List[T] {
	ifb := &List[T]{
		cfg:           cfg,
		parent:        parent,
		keymaps:       []Keymap{makeKeymap[T]()},
		UpdateList:    updateList,
		OpenSelection: openSelection,
		ItemRepr:      repr,
	}
	if initialList != nil {
		iList := initialList()
		ifb.Items = iList
	}

	ifb.keymaps = append(ifb.keymaps, MakeInsertionKeys(func(c *Context, b byte) error {
		return ifb.InsertCharAtBuffer(b)
	}))
	return ifb
}

func (l *List[T]) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if l.LastInputWeRanUpdateFor != string(l.UserInput) {
		l.LastInputWeRanUpdateFor = string(l.UserInput)
		l.UpdateList(l, string(l.UserInput))
	}
	charSize := measureTextSize(l.parent.Font, ' ', l.parent.FontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y)*2, l.cfg.CurrentThemeColors().StatusBarBackground.ToColorRGBA())
	rl.DrawTextEx(l.parent.Font, string(l.UserInput), rl.Vector2{
		X: zeroLocation.X, Y: zeroLocation.Y + charSize.Y/2,
	}, float32(l.parent.FontSize), 0, l.cfg.CurrentThemeColors().Foreground.ToColorRGBA())

	switch l.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(float64(zeroLocation.X)+float64(charSize.X))*int32(l.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(float64(zeroLocation.X)+float64(charSize.X))*int32(l.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(float64(zeroLocation.X)+float64(charSize.X))*int32(l.Idx), int32(zeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(zeroLocation.Y) + int32(3*(charSize.Y))
	maxLine := int(int32((maxH+float64(zeroLocation.Y))-float64(startOfListY)) / int32(charSize.Y))

	//draw list of items
	for idx, item := range l.VisibleView(maxLine) {
		rl.DrawTextEx(l.parent.Font, l.ItemRepr(item), rl.Vector2{
			X: zeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, float32(l.parent.FontSize), 0, l.cfg.CurrentThemeColors().Foreground.ToColorRGBA())
	}
	if len(l.Items) > 0 {
		rl.DrawRectangle(int32(zeroLocation.X), int32(int(startOfListY)+(l.Selection-l.VisibleStart)*int(charSize.Y)), int32(maxW), int32(charSize.Y), rl.Fade(l.cfg.CurrentThemeColors().SelectionBackground.ToColorRGBA(), 0.2))
	}
}

func makeKeymap[T any]() Keymap {
	return Keymap{

		Key{K: "f", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.CursorRight(1)
		}),
		Key{K: "v", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.Paste()
		}),
		Key{K: "c", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.Copy()
		}),
		Key{K: "a", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.BeginningOfTheLine()
		}),
		Key{K: "e", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.EndOfTheLine()
		}),
		Key{K: "g", Control: true}: MakeCommand(func(e *List[T]) error {
			e.parent.KillDrawable(e.ID)
			return nil
		}),

		Key{K: "<right>"}: MakeCommand(func(e *List[T]) error {
			return e.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.NextWordStart()
		}),
		Key{K: "<left>"}: MakeCommand(func(e *List[T]) error {
			return e.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.PreviousWord()
		}),

		Key{K: "p", Control: true}: MakeCommand(func(e *List[T]) error {
			e.PrevItem()
			return nil
		}),
		Key{K: "n", Control: true}: MakeCommand(func(e *List[T]) error {
			e.NextItem()
			return nil
		}),
		Key{K: "<up>"}: MakeCommand(func(e *List[T]) error {
			e.PrevItem()

			return nil
		}),
		Key{K: "<down>"}: MakeCommand(func(e *List[T]) error {
			e.NextItem()
			return nil
		}),
		Key{K: "b", Control: true}: MakeCommand(func(e *List[T]) error {
			return e.CursorLeft(1)
		}),
		Key{K: "<home>"}: MakeCommand(func(e *List[T]) error {
			return e.BeginningOfTheLine()
		}),

		Key{K: "<enter>"}: MakeCommand(func(e *List[T]) error {
			if len(e.Items) > 0 && len(e.Items) > e.Selection {
				return e.OpenSelection(e.parent, e.Items[e.Selection])
			}

			return nil
		}),
		Key{K: "<backspace>"}:                MakeCommand(func(e *List[T]) error { return e.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: MakeCommand(func(e *List[T]) error { return e.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           MakeCommand(func(e *List[T]) error { return e.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               MakeCommand(func(e *List[T]) error { return e.DeleteWordForward() }),
		Key{K: "<delete>"}:                   MakeCommand(func(e *List[T]) error { return e.DeleteCharForward() }),
	}
}

func NewBufferList(parent *Context, cfg *Config) *List[ScoredItem[Drawable]] {
	updateList := func(l *List[ScoredItem[Drawable]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, fmt.Sprint(item.Item))
		}

		sortme(l.Items, func(t1 ScoredItem[Drawable], t2 ScoredItem[Drawable]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Context, item ScoredItem[Drawable]) error {
		parent.KillDrawable(parent.ActiveDrawable().GetID())
		parent.MarkDrawableAsActive(item.Item.GetID())

		return nil
	}
	initialList := func() []ScoredItem[Drawable] {
		var buffers []ScoredItem[Drawable]
		for _, v := range parent.Drawables {
			if v != nil {
				buffers = append(buffers, ScoredItem[Drawable]{Item: v})
			}
		}

		return buffers
	}
	repr := func(s ScoredItem[Drawable]) string {
		return s.Item.String()
	}
	return NewList[ScoredItem[Drawable]](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		initialList,
	)

}

func NewThemeList(parent *Context, cfg *Config) *List[ScoredItem[string]] {
	updateList := func(l *List[ScoredItem[string]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, fmt.Sprint(item.Item))
		}

		sortme(l.Items, func(t1 ScoredItem[string], t2 ScoredItem[string]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Context, item ScoredItem[string]) error {
		parent.Cfg.CurrentTheme = item.Item
		parent.KillDrawable(parent.ActiveDrawableID())
		return nil
	}
	initialList := func() []ScoredItem[string] {
		var themes []ScoredItem[string]
		for _, v := range parent.Cfg.Themes {
			themes = append(themes, ScoredItem[string]{Item: v.Name})
		}

		return themes
	}
	repr := func(s ScoredItem[string]) string {
		return s.Item
	}
	return NewList[ScoredItem[string]](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		initialList,
	)

}

type GrepLocationItem struct {
	Filename string
	Text     string
	Line     int
	Col      int
}

type LocationItem struct {
	Filename string
}

func NewFuzzyFileList(parent *Context, cfg *Config, cwd string) *List[ScoredItem[LocationItem]] {
	updateList := func(l *List[ScoredItem[LocationItem]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, item.Item.Filename)
		}

		sortme(l.Items, func(t1 ScoredItem[LocationItem], t2 ScoredItem[LocationItem]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Context, item ScoredItem[LocationItem]) error {
		err := SwitchOrOpenFileInCurrentWindow(parent, parent.Cfg, path.Join(cwd, item.Item.Filename), nil)
		if err != nil {
			panic(err)
		}
		return nil
	}

	repr := func(g ScoredItem[LocationItem]) string {
		return fmt.Sprintf("%s", g.Item.Filename)
	}

	initialList := func() []ScoredItem[LocationItem] {
		var locationItems []ScoredItem[LocationItem]
		files := RipgrepFiles(cwd)
		for _, file := range files {
			locationItems = append(locationItems, ScoredItem[LocationItem]{Item: LocationItem{Filename: file}})
		}

		return locationItems

	}

	return NewList[ScoredItem[LocationItem]](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		initialList,
	)
}

func NewFileList(parent *Context, cfg *Config, initialInput string) *List[LocationItem] {
	updateList := func(l *List[LocationItem], input string) {
		matches, err := filepath.Glob(string(input) + "*")
		if err != nil {
			return
		}

		l.Items = nil

		for _, match := range matches {
			stat, err := os.Stat(match)
			if err == nil {
				isDir := stat.IsDir()
				_ = isDir
			}
			l.Items = append(l.Items, LocationItem{
				Filename: match,
			})
		}

		if l.Selection >= len(l.Items) {
			l.Selection = len(l.Items) - 1
		}

		if l.Selection < 0 {
			l.Selection = 0
		}

		return

	}
	openUserInput := func(parent *Context, userInput string) {
		parent.KillDrawable(parent.ActiveDrawableID())
		err := SwitchOrOpenFileInCurrentWindow(parent, parent.Cfg, userInput, nil)
		if err != nil {
			panic(err)
		}
	}
	openSelection := func(parent *Context, item LocationItem) error {
		parent.KillDrawable(parent.ActiveDrawableID())
		err := SwitchOrOpenFileInCurrentWindow(parent, parent.Cfg, item.Filename, nil)
		if err != nil {
			panic(err)
		}
		return nil
	}

	repr := func(g LocationItem) string {
		return fmt.Sprintf("%s", g.Filename)
	}

	tryComplete := func(f *List[LocationItem]) error {
		input := f.UserInput

		matches, err := filepath.Glob(string(input) + "*")
		if err != nil {
			return nil
		}

		if f.Selection < len(f.Items) {
			stat, err := os.Stat(matches[f.Selection])
			if err == nil {
				if stat.IsDir() {
					matches[f.Selection] += "/"
				}
			}
			f.UserInput = []byte(matches[f.Selection])
			f.CursorRight(len(f.UserInput) - len(input))
		}
		return nil
	}

	ifb := NewList[LocationItem](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		nil,
	)

	ifb.keymaps[0][Key{K: "<enter>", Control: true}] = func(preditor *Context) error {
		input := preditor.ActiveDrawable().(*List[LocationItem]).UserInput
		openUserInput(preditor, string(input))
		return nil
	}
	ifb.keymaps[0][Key{K: "<tab>"}] = MakeCommand(tryComplete)
	var absRoot string
	var err error
	if initialInput == "" {
		absRoot, _ = os.Getwd()
	} else {
		absRoot, err = filepath.Abs(initialInput)
		if err != nil {
			panic(err)
		}
	}
	ifb.SetNewUserInput([]byte(absRoot))

	return ifb
}
