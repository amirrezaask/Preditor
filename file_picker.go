package main


type FilePicker struct {
	Root string
	keymaps []Keymap
	maxHeight int32
	maxWidth int32
}

func (f *FilePicker) Render(){

}

func (f *FilePicker) SetMaxWidth(w int32){
	f.maxWidth = w
}

func (f *FilePicker) SetMaxHeight(h int32){
	f.maxHeight = h
}

func (f *FilePicker) GetMaxWidth() int32{
	return f.maxWidth
}

func (f *FilePicker) GetMaxHeight() int32{
	return f.maxHeight
}

func (f *FilePicker) Keymaps() []Keymap{
	return f.keymaps
}

