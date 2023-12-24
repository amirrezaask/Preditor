# Preditor
## Programmable Editor
Simple text editor implemented in Golang using Raylib with the goal of being a simple, extensible basic editor that you can build your own PDE on.
 
## Features So far

- reading/writing files
- simple keybindings (no key chords yet)
- left mouse click
- mouse scroll
- basic emacs bindings for movements
- Line numbers
- Basic status bar to show file name, file state
- Selecting text
- Incremental Buffer Search
- Support BeforeSaveHook to use formatting tools like Go fmt
- Auto format go code using pkg/format


## TODO:
- Zoom in/out (increase/decrease) font size
- Copy/paste
- Undo
- Fuzzy file finder
- Command output buffer ( run a command and see it's result, similar to *Compile Mode* in Emacs)

# Screenshot
![Main.go](assets/screenshot.png)
