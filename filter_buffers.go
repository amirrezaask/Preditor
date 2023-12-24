package preditor

import (
	"fmt"
	"github.com/amirrezaask/preditor/components"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"os"
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
	parent                  *Preditor
	keymaps                 []Keymap
	List                    components.ListComponent[T]
	UserInputComponent      *components.UserInputComponent
	LastInputWeRanUpdateFor string
	UpdateList              func(list *components.ListComponent[T], input string)
	OpenSelection           func(preditor *Preditor, t T) error
	ItemRepr                func(item T) string
}

func (i InteractiveFilterBuffer[T]) Keymaps() []Keymap {
	return i.keymaps
}

func (i InteractiveFilterBuffer[T]) String() string {
	return fmt.Sprintf("InteractiveFilterBuffer: %T", *new(T))
}

func NewInteractiveFilterBuffer[T any](
	parent *Preditor,
	cfg *Config,
	updateList func(list *components.ListComponent[T], input string),
	openSelection func(preditor *Preditor, t T) error,
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
	return ifb
}

func (i *InteractiveFilterBuffer[T]) Render(zeroLocation rl.Vector2, maxH int32, maxW int32) {
	if i.LastInputWeRanUpdateFor != string(i.UserInputComponent.UserInput) {
		i.LastInputWeRanUpdateFor = string(i.UserInputComponent.UserInput)
		i.UpdateList(&i.List, string(i.UserInputComponent.UserInput))
	}
	charSize := measureTextSize(i.parent.Font, ' ', i.parent.FontSize, 0)

	//draw input box
	rl.DrawRectangleLines(int32(zeroLocation.X), int32(zeroLocation.Y), maxW, int32(charSize.Y)*2, i.cfg.Colors.StatusBarBackground)
	rl.DrawTextEx(i.parent.Font, string(i.UserInputComponent.UserInput), rl.Vector2{
		X: zeroLocation.X, Y: zeroLocation.Y + charSize.Y/2,
	}, float32(i.parent.FontSize), 0, i.cfg.Colors.Foreground)

	switch i.cfg.CursorShape {
	case CURSOR_SHAPE_OUTLINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(i.UserInputComponent.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_BLOCK:
		rl.DrawRectangle(int32(charSize.X)*int32(i.UserInputComponent.Idx), int32(zeroLocation.Y+charSize.Y/2), int32(charSize.X), int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	case CURSOR_SHAPE_LINE:
		rl.DrawRectangleLines(int32(charSize.X)*int32(i.UserInputComponent.Idx), int32(zeroLocation.Y+charSize.Y/2), 2, int32(charSize.Y), rl.Fade(rl.Red, 0.5))
	}

	startOfListY := int32(zeroLocation.Y) + int32(3*(charSize.Y))
	maxLine := int((maxH - startOfListY) / int32(charSize.Y))

	//draw list of items
	for idx, item := range i.List.VisibleView(maxLine) {
		rl.DrawTextEx(i.parent.Font, i.ItemRepr(item), rl.Vector2{
			X: zeroLocation.X, Y: float32(startOfListY) + float32(idx)*charSize.Y,
		}, float32(i.parent.FontSize), 0, i.cfg.Colors.Foreground)
	}
	if len(i.List.Items) > 0 {
		rl.DrawRectangle(int32(zeroLocation.X), int32(int(startOfListY)+(i.List.Selection-i.List.VisibleStart)*int(charSize.Y)), maxW, int32(charSize.Y), rl.Fade(i.cfg.Colors.Selection, 0.2))
	}
}

func MakeInteractiveFilterBufferCommand[T any](f func(e *InteractiveFilterBuffer[T]) error) Command {
	return func(preditor *Preditor) error {
		defer handlePanicAndWriteMessage(preditor)
		buffer := preditor.ActiveBuffer().(*InteractiveFilterBuffer[T])
		return f(buffer)
	}
}

func makeKeymap[T any]() Keymap {
	return Keymap{
		Key{K: "=", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.parent.IncreaseFontSize(5)

			return nil
		}),
		Key{K: "-", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.parent.DecreaseFontSize(5)

			return nil
		}),
		Key{K: "f", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "v", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.Paste()
		}),
		Key{K: "c", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.Copy()
		}),
		Key{K: "s", Control: true}: MakeInteractiveFilterBufferCommand(func(a *InteractiveFilterBuffer[T]) error {
			a.keymaps = append(a.keymaps, SearchTextBufferKeymap)
			return nil
		}),

		Key{K: "a", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.BeginningOfTheLine()
		}),
		Key{K: "e", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.EndOfTheLine()
		}),

		Key{K: "<right>"}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorRight(1)
		}),
		Key{K: "<right>", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.NextWordStart()
		}),
		Key{K: "<left>"}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<left>", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.PreviousWord()
		}),

		Key{K: "p", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.PrevItem()
			return nil
		}),
		Key{K: "n", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "<up>"}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.PrevItem()

			return nil
		}),
		Key{K: "<down>"}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			e.List.NextItem()
			return nil
		}),
		Key{K: "b", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.CursorLeft(1)
		}),
		Key{K: "<home>"}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.UserInputComponent.BeginningOfTheLine()
		}),

		//insertion
		Key{K: "<enter>"}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error {
			return e.OpenSelection(e.parent, e.List.Items[e.List.Selection])
		}),
		Key{K: "<space>"}:                    MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.InsertCharAtBuffer(' ') }),
		Key{K: "<backspace>"}:                MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteCharBackward() }),
		Key{K: "<backspace>", Control: true}: MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteWordBackward() }),
		Key{K: "d", Control: true}:           MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "d", Alt: true}:               MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteWordForward() }),
		Key{K: "<delete>"}:                   MakeInteractiveFilterBufferCommand(func(e *InteractiveFilterBuffer[T]) error { return e.UserInputComponent.DeleteCharForward() }),
		Key{K: "a"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('a') }),
		Key{K: "b"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('b') }),
		Key{K: "c"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('c') }),
		Key{K: "d"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('d') }),
		Key{K: "e"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('e') }),
		Key{K: "f"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('f') }),
		Key{K: "g"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('g') }),
		Key{K: "h"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('h') }),
		Key{K: "i"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('i') }),
		Key{K: "j"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('j') }),
		Key{K: "k"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('k') }),
		Key{K: "l"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('l') }),
		Key{K: "m"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('m') }),
		Key{K: "n"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('n') }),
		Key{K: "o"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('o') }),
		Key{K: "p"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('p') }),
		Key{K: "q"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('q') }),
		Key{K: "r"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('r') }),
		Key{K: "s"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('s') }),
		Key{K: "t"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('t') }),
		Key{K: "u"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('u') }),
		Key{K: "v"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('v') }),
		Key{K: "w"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('w') }),
		Key{K: "x"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('x') }),
		Key{K: "y"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('y') }),
		Key{K: "z"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('z') }),
		Key{K: "0"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('0') }),
		Key{K: "1"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('1') }),
		Key{K: "2"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('2') }),
		Key{K: "3"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('3') }),
		Key{K: "4"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('4') }),
		Key{K: "5"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('5') }),
		Key{K: "6"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('6') }),
		Key{K: "7"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('7') }),
		Key{K: "8"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('8') }),
		Key{K: "9"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('9') }),
		Key{K: "\\"}:                         MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('\\') }),
		Key{K: "\\", Shift: true}:            MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('|') }),
		Key{K: "0", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer(')') }),
		Key{K: "1", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('!') }),
		Key{K: "2", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('@') }),
		Key{K: "3", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('#') }),
		Key{K: "4", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('$') }),
		Key{K: "5", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('%') }),
		Key{K: "6", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('^') }),
		Key{K: "7", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('&') }),
		Key{K: "8", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('*') }),
		Key{K: "9", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('(') }),
		Key{K: "a", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('A') }),
		Key{K: "b", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('B') }),
		Key{K: "c", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('C') }),
		Key{K: "d", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('D') }),
		Key{K: "e", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('E') }),
		Key{K: "f", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('F') }),
		Key{K: "g", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('G') }),
		Key{K: "h", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('H') }),
		Key{K: "i", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('I') }),
		Key{K: "j", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('J') }),
		Key{K: "k", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('K') }),
		Key{K: "l", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('L') }),
		Key{K: "m", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('M') }),
		Key{K: "n", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('N') }),
		Key{K: "o", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('O') }),
		Key{K: "p", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('P') }),
		Key{K: "q", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('Q') }),
		Key{K: "r", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('R') }),
		Key{K: "s", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('S') }),
		Key{K: "t", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('T') }),
		Key{K: "u", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('U') }),
		Key{K: "v", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('V') }),
		Key{K: "w", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('W') }),
		Key{K: "x", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('X') }),
		Key{K: "y", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('Y') }),
		Key{K: "z", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('Z') }),
		Key{K: "["}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('[') }),
		Key{K: "]"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer(']') }),
		Key{K: "[", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('{') }),
		Key{K: "]", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('}') }),
		Key{K: ";"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer(';') }),
		Key{K: ";", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer(':') }),
		Key{K: "'"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('\'') }),
		Key{K: "'", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('"') }),
		Key{K: "\""}:                         MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('"') }),
		Key{K: ","}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer(',') }),
		Key{K: "."}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('.') }),
		Key{K: ",", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('<') }),
		Key{K: ".", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('>') }),
		Key{K: "/"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('/') }),
		Key{K: "/", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('?') }),
		Key{K: "-"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('-') }),
		Key{K: "="}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('=') }),
		Key{K: "-", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('_') }),
		Key{K: "=", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('+') }),
		Key{K: "`"}:                          MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('`') }),
		Key{K: "`", Shift: true}:             MakeInteractiveFilterBufferCommand(func(f *InteractiveFilterBuffer[T]) error { return f.UserInputComponent.InsertCharAtBuffer('~') }),
	}
}

func NewBufferSwitcher(parent *Preditor, cfg *Config) *InteractiveFilterBuffer[ScoredItem[Buffer]] {
	updateList := func(l *components.ListComponent[ScoredItem[Buffer]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, fmt.Sprint(item.Item))
		}

		sortme(l.Items, func(t1 ScoredItem[Buffer], t2 ScoredItem[Buffer]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Preditor, item ScoredItem[Buffer]) error {
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

type GrepLocationItem struct {
	Filename string
	Text     string
	Line     int
	Col      int
}

func NewGrepBuffer(
	parent *Preditor,
	cfg *Config,

) *InteractiveFilterBuffer[GrepLocationItem] {
	updateList := func(l *components.ListComponent[GrepLocationItem], input string) {
		if len(input) < 3 {
			return
		}
		c := RipgrepAsync(string(input))
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
	openSelection := func(parent *Preditor, item GrepLocationItem) error {
		return SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, item.Filename, &Position{Line: item.Line, Column: item.Col})
	}

	repr := func(g GrepLocationItem) string {
		return fmt.Sprintf("%s:%d: %s", g.Filename, g.Line, g.Text)
	}

	return NewInteractiveFilterBuffer[GrepLocationItem](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		nil,
	)
}

type LocationItem struct {
	Filename string
}

func NewFuzzyFileBuffer(parent *Preditor, cfg *Config) *InteractiveFilterBuffer[ScoredItem[LocationItem]] {
	updateList := func(l *components.ListComponent[ScoredItem[LocationItem]], input string) {
		for idx, item := range l.Items {
			l.Items[idx].Score = fuzzy.RankMatchNormalizedFold(input, item.Item.Filename)
		}

		sortme(l.Items, func(t1 ScoredItem[LocationItem], t2 ScoredItem[LocationItem]) bool {
			return t1.Score > t2.Score
		})

	}
	openSelection := func(parent *Preditor, item ScoredItem[LocationItem]) error {
		err := SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, item.Item.Filename, nil)
		if err != nil {
			panic(err)
		}
		return nil
	}

	repr := func(g ScoredItem[LocationItem]) string {
		return fmt.Sprintf("%s", g.Item.Filename)
	}

	return NewInteractiveFilterBuffer[ScoredItem[LocationItem]](
		parent,
		cfg,
		updateList,
		openSelection,
		repr,
		nil,
	)
}

func NewFilePickerBuffer(parent *Preditor, cfg *Config, initialInput string) *InteractiveFilterBuffer[LocationItem] {
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
	openUserInput := func(parent *Preditor, userInput string) {
		parent.KillBuffer(parent.ActiveBufferID)
		err := SwitchOrOpenFileInTextBuffer(parent, parent.Cfg, userInput, nil)
		if err != nil {
			panic(err)
		}
	}
	openSelection := func(parent *Preditor, item LocationItem) error {
		parent.KillBuffer(parent.ActiveBufferID)
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

	ifb.keymaps[0][Key{K: "<enter>", Control: true}] = func(preditor *Preditor) error {
		input := preditor.ActiveBuffer().(*InteractiveFilterBuffer[LocationItem]).UserInputComponent.UserInput
		openUserInput(preditor, string(input))
		return nil
	}
	ifb.keymaps[0][Key{K: "<tab>"}] = MakeInteractiveFilterBufferCommand(tryComplete)
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
