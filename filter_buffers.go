package preditor

import (
	"fmt"
	"github.com/amirrezaask/preditor/components"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type ScoredItem[T any] struct {
	Item  T
	Score int
}

type InteractiveFilterBuffer[T any] struct {
	BaseBuffer
	cfg                     *Config
	parent                  *Context
	keymaps                 []Keymap
	List                    components.ListComponent[T]
	UserInputComponent      *components.UserInputComponent
	LastInputWeRanUpdateFor string
	UpdateList              func(list *components.ListComponent[T], input string)
	OpenSelection           func(preditor *Context, t T) error
	ItemRepr                func(item T) string
}

func (i InteractiveFilterBuffer[T]) Keymaps() []Keymap {
	return i.keymaps
}

func (i InteractiveFilterBuffer[T]) String() string {
	return fmt.Sprintf("InteractiveFilterBuffer: %T", *new(T))
}

func NewInteractiveFilterBuffer[T any](
	parent *Context,
	cfg *Config,
	updateList func(list *components.ListComponent[T], input string),
	openSelection func(preditor *Context, t T) error,
	repr func(t T) string,
	initialList func() []T,
) *InteractiveFilterBuffer[T] {
	ifb := &InteractiveFilterBuffer[T]{
		cfg:                cfg,
		parent:             parent,
		keymaps:            []Keymap{makeKeymap[T]()},
		UserInputComponent: components.NewUserInputComponent(),
		UpdateList:         updateList,
		OpenSelection:      openSelection,
		ItemRepr:           repr,
	}
	if initialList != nil {
		iList := initialList()
		ifb.List.Items = iList
	}

	ifb.keymaps = append(ifb.keymaps, MakeInsertionKeys(func(c *Context, b byte) error {
		return ifb.UserInputComponent.InsertCharAtBuffer(b)
	}))
	return ifb
}

func (i *InteractiveFilterBuffer[T]) Render(zeroLocation rl.Vector2, maxH float64, maxW float64) {
	if i.LastInputWeRanUpdateFor != string(i.UserInputComponent.UserInput) {
		i.LastInputWeRanUpdateFor = string(i.UserInputComponent.UserInput)
		i.UpdateList(&i.List, string(i.UserInputComponent.UserInput))
	}
	charSize := measureTextSize(i.parent.Font, ' ', i.parent.FontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(zeroLocation.X), int32(zeroLocation.Y), int32(maxW), int32(charSize.Y)*2, i.cfg.CurrentThemeColors().StatusBarBackground)
	rl.DrawTextEx(i.parent.Font, string(i.UserInputComponent.UserInput), rl.Vector2{
		X: zeroLocation.X, Y: zeroLocation.Y + charSize.Y/2,
	}, float32(i.parent.FontSize), 0, i.cfg.CurrentThemeColors().Foreground)

	switch i.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(float64(zeroLocation.X)+float64(charSize.X))*int32(i.UserInputComponent.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(float64(zeroLocation.X)+float64(charSize.X))*int32(i.UserInputComponent.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(float64(zeroLocation.X)+float64(charSize.X))*int32(i.UserInputComponent.Idx), int32(zeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(zeroLocation.Y) + int32(3*(charSize.Y))
	maxLine := int(int32(maxH-float64(startOfListY)) / int32(charSize.Y))

	//draw list of items
	for idx, item := range i.List.VisibleView(maxLine) {
		rl.DrawTextEx(i.parent.Font, i.ItemRepr(item), rl.Vector2{
			X: zeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, float32(i.parent.FontSize), 0, i.cfg.CurrentThemeColors().Foreground)
	}
	if len(i.List.Items) > 0 {
		rl.DrawRectangle(int32(zeroLocation.X), int32(int(startOfListY)+(i.List.Selection-i.List.VisibleStart)*int(charSize.Y)), int32(maxW), int32(charSize.Y), rl.Fade(i.cfg.CurrentThemeColors().Selection, 0.2))
	}
}

func makeKeymap[T any]() Keymap {
	return Keymap{

		Key{K: "f", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "v", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.Paste()
		}),
		Key{K: "c", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.Copy()
		}),
		Key{K: "a", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.BeginningOfTheLine()
		}),
		Key{K: "e", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.EndOfTheLine()
		}),

		Key{K: "<right>"}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.NextWordStart()
		}),
		Key{K: "<left>"}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.PreviousWord()
		}),

		Key{K: "p", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.PrevItem()
			return nil
		}),
		Key{K: "n", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "<up>"}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.PrevItem()

			return nil
		}),
		Key{K: "<down>"}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "b", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<home>"}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.BeginningOfTheLine()
		}),

		Key{K: "<enter>"}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.OpenSelection(e.parent, e.List.Items[e.List.Selection])
		}),
		Key{K: "<backspace>"}:                MakeCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: MakeCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           MakeCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               MakeCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteWordForward() }),
		Key{K: "<delete>"}:                   MakeCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteCharForward() }),
	}
}

func NewBufferSwitcher(parent *Context, cfg *Config) *InteractiveFilterBuffer[ScoredItem[Buffer]] {
	updateList := func(l *components.ListComponent[ScoredItem[Buffer]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, fmt.Sprint(item.Item))
		}

		sortme(l.Items, func(t1 ScoredItem[Buffer], t2 ScoredItem[Buffer]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Context, item ScoredItem[Buffer]) error {
		parent.KillBuffer(parent.ActiveBuffer().GetID())
		parent.MarkBufferAsActive(item.Item.GetID())

		return nil
	}
	initialList := func() []ScoredItem[Buffer] {
		var buffers []ScoredItem[Buffer]
		for _, v := range parent.Buffers {
			buffers = append(buffers, ScoredItem[Buffer]{Item: v})
		}

		return buffers
	}
	repr := func(s ScoredItem[Buffer]) string {
		return s.Item.String()
	}
	return NewInteractiveFilterBuffer[ScoredItem[Buffer]](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		initialList,
	)

}

func NewThemeSwitcher(parent *Context, cfg *Config) *InteractiveFilterBuffer[ScoredItem[string]] {
	updateList := func(l *components.ListComponent[ScoredItem[string]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, fmt.Sprint(item.Item))
		}

		sortme(l.Items, func(t1 ScoredItem[string], t2 ScoredItem[string]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Context, item ScoredItem[string]) error {
		parent.Cfg.CurrentTheme = item.Item
		parent.KillBuffer(parent.ActiveBufferID())
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
	return NewInteractiveFilterBuffer[ScoredItem[string]](
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

func NewGrepBuffer(
	parent *Context,
	cfg *Config,
	cwd string,

) *InteractiveFilterBuffer[GrepLocationItem] {
	updateList := func(l *components.ListComponent[GrepLocationItem], input string) {
		if len(input) < 3 {
			return
		}
		c := RipgrepAsync(string(input), cwd)
		go func() {
			lines := <-c
			l.Items = nil

			for _, line := range lines {
				lineS := string(line)
				segs := strings.SplitN(lineS, ":", 4)
				if len(segs) < 4 {
					continue
				}

				line, err := strconv.Atoi(segs[1])
				if err != nil {
					continue
				}
				col, err := strconv.Atoi(segs[2])
				if err != nil {
					continue
				}
				l.Items = append(l.Items, GrepLocationItem{
					Filename: segs[0],
					Line:     line,
					Col:      col,
					Text:     segs[3],
				})
			}
		}()

	}
	openSelection := func(parent *Context, item GrepLocationItem) error {
		return SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, path.Join(cwd, item.Filename), &Position{Line: item.Line, Column: item.Col})
	}

	repr := func(g GrepLocationItem) string {
		return fmt.Sprintf("%s:%d: %s", g.Filename, g.Line, g.Text)
	}

	gb := NewInteractiveFilterBuffer[GrepLocationItem](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		nil,
	)

	//gb.keymaps[0].SetKey(Key{K: "x", Control: true}, func(c *Context) error {
	//TODO(amirreza): export results into a text buffer

	//return nil
	// })

	return gb
}

type LocationItem struct {
	Filename string
}

func NewFuzzyFileBuffer(parent *Context, cfg *Config, cwd string) *InteractiveFilterBuffer[ScoredItem[LocationItem]] {
	updateList := func(l *components.ListComponent[ScoredItem[LocationItem]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, item.Item.Filename)
		}

		sortme(l.Items, func(t1 ScoredItem[LocationItem], t2 ScoredItem[LocationItem]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Context, item ScoredItem[LocationItem]) error {
		err := SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, path.Join(cwd, item.Item.Filename), nil)
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

	return NewInteractiveFilterBuffer[ScoredItem[LocationItem]](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		initialList,
	)
}

func NewFilePickerBuffer(parent *Context, cfg *Config, initialInput string) *InteractiveFilterBuffer[LocationItem] {
	updateList := func(l *components.ListComponent[LocationItem], input string) {
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
		parent.KillBuffer(parent.ActiveBufferID())
		err := SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, userInput, nil)
		if err != nil {
			panic(err)
		}
	}
	openSelection := func(parent *Context, item LocationItem) error {
		parent.KillBuffer(parent.ActiveBufferID())
		err := SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, item.Filename, nil)
		if err != nil {
			panic(err)
		}
		return nil
	}

	repr := func(g LocationItem) string {
		return fmt.Sprintf("%s", g.Filename)
	}

	tryComplete := func(f *InteractiveFilterBuffer[LocationItem]) error {
		input := f.UserInputComponent.UserInput

		matches, err := filepath.Glob(string(input) + "*")
		if err != nil {
			return nil
		}

		if len(matches) == 1 {
			stat, err := os.Stat(matches[0])
			if err == nil {
				if stat.IsDir() {
					matches[0] += "/"
				}
			}
			f.UserInputComponent.UserInput = []byte(matches[0])
			f.UserInputComponent.CursorRight(len(f.UserInputComponent.UserInput) - len(input))
		}
		return nil
	}

	ifb := NewInteractiveFilterBuffer[LocationItem](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		nil,
	)

	ifb.keymaps[0][Key{K: "<enter>", Control: true}] = func(preditor *Context) error {
		input := preditor.ActiveBuffer().(*InteractiveFilterBuffer[LocationItem]).UserInputComponent.UserInput
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
	ifb.UserInputComponent.SetNewUserInput([]byte(absRoot))

	return ifb
}
