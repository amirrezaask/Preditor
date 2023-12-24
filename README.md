# Preditor
## Programmable Editor
Simple text editor implemented in Golang using Raylib with the goal of replacing Emacs for me, easier to extend and much faster and better language to work with than Elisp.

# Screenshots
![Main](assets/main.png)
#### File Picker
![Main](assets/file-picker.png)
#### Searching text (ripgrep backend)
![Main](assets/search-grep.png)
#### Split windows
![Main](assets/split-windows.png)
#### Build window
![Main](assets/build-window.png)
![Main](assets/build-window-max.png)

## Features
- reading/writing files
- simple keybindings (no key chords yet)
- Scrolling with both mouse and keyboard
- Line numbers
- Statusbar
- Selecting text
- Cut/Copy/Paste
- Incremental Search
- Multi-Cursors (WIP)
- Support BeforeSaveHook to use formatting tools like Go fmt
- Auto format go code using pkg/format
- Basic Regex syntax highlighting for keywords and types.
- Undo
- Open file with glob completion
- switch between open files
- Grep Buffers using Ripgrep ( more backends are possible )
- Fuzzy file finder
- Zoom in/out (increase/decrease) font size
- Multi Window ( Splits )
- Compile commands

