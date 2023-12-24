# Preditor
## Programmable Editor
Simple text editor implemented in Golang using Raylib with the goal of replacing Emacs for me, easier to extend and much faster and better language to work with than Elisp.

## Show case
![Opening Files](assets/file-opening.gif)
![Opening Files](assets/searching.gif)
![Opening Files](assets/moving-around.gif)

# Demo
[![Demo](http://img.youtube.com/vi/ogmozlzDAPY/0.jpg)](http://www.youtube.com/watch?v=ogmozlzDAPY)
 
## Features

- reading/writing files
- simple keybindings (no key chords yet)
- Scrolling with both mouse and keyboard
- Line numbers
- Statusbar
- Selecting text
- Cut/Copy/Paste
- Incremental Search
- Support BeforeSaveHook to use formatting tools like Go fmt
- Auto format go code using pkg/format
- Basic Regex syntax highlighting for keywords and types.
- Undo (still work in progress!)
- Open file with glob completion
- switch between open files


## TODO:
- Zoom in/out (increase/decrease) font size
- Fuzzy file finder
- Command output buffer ( run a command and see it's result, similar to *Compile Mode* in Emacs)

# Screenshot
![Main.go](assets/screenshot.png)
![Open File Menu](assets/files.png)
