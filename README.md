# Compromise (and Bash/Zsh completion for ADB and Go)

Compromise is a Go framework for writing shell completion for Bash / Zsh.

Currently it comes with the following completion:

 - ADB (Android Debug Bridge), fastboot and atest
 - Go


## Caveat
 It's still in an alpha stage. Details are subject to change, but feedback is welcome.

## Installing ADB and/or Go completion

Assuming you have the `go` command installed, just run the following commands in your shell.
```bash
# Get and install the binaries.
go get -u github.com/omakoto/compromise/src/cmds/...

# Then to install completion, run them on your shell.
# Add them to your shell's RC file (.bashrc or .zshrc) to make them persistent.
. <(compromise-adb) # Install ADB / fastboot / atest completion
. <(compromise-go)  # Install Go completion
```
 
 *NOTE `go run` won't work; you need to actually compile them.*
 
### Creating aliases to direct dumpsys/etc 
 - `compromise-adb` also installs completion for some "shortcut" commands,
so if you have following aliases, completion will work for them too.

```bash
alias dumpsys="adb shell dumpsys"
alias cmd="adb shell cmd"
alias am="adb shell am"
alias pm="adb shell pm"
  :
```  
For the full supported command name list, see [the source code](src/cmds/compromise-adb/adb.go).
 
 - If you do not want to install completion for all the listed commands
   in the source file, pass the command name you want to use as arguments. Example: 

```bash
. <(compromise-adb adb dumpsys) # Only install competion for the adb command.  
```

## Features

 - Define completion in [a simple-ish language](src/cmds/compromise-go/go.go) that supports both Bash and Zsh.
 - Generate candidates with [custom Go functions](src/cmds/compromise-adb/adb.go).
 - Show candidate description not only on Zsh but on Bash too.
   - On Bash, completion candidates look like this (type `adb[SPACE][TAB]`):
 <img src="https://raw.githubusercontent.com/omakoto/compromise/master/img/compromise-adb.png" width=600>
 
## Known Issues

 - Not heavily tested on Zsh yet.

## TODOs
 - Write tests
