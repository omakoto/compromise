[![Build Status](https://travis-ci.org/omakoto/compromise.svg?branch=master)](https://travis-ci.org/omakoto/compromise)
# Compromise (and Bash/Zsh completion for ADB and Go)

Compromise is a Go framework for writing shell completion for Bash / Zsh.

Currently it comes with the following two sets of completions:

 - ADB (Android Debug Bridge, including several shell commands), fastboot, atest, m/mm/mmm, etc
    - Examples:
        - Most `adb` subcommands, major subcommands for `adb shell [am|pm|settings]`
         (e.g. `am start-activity [flags] INTENT`), service name for `adb shell [dumpsys|cmd]`
        - Flags for `fastboot` 
        - Build target and some pseudo build targets for `m` (`m MODULE`, `m installclean`, etc)
        - `atest MODULE`, `atest FILENAME#method1,method2,...`
        - Other commands such as `runahat PROCESSNAME|PID` and `stacks PROCESSNAME|PID`
 - Go

## Caveat
 It's still in an alpha stage. Details are subject to change, but feedback is welcome.

## Features

 - Define completion in [an obvious-ish language](src/cmds/compromise-go/go.go) that supports both Bash and Zsh.
 - Generate dynamic candidates with [custom Go functions](src/cmds/compromise-adb/adb.go).
 - Show candidate description not only on Zsh but on Bash too.
   - On Bash, completion candidates look like this (type `adb[SPACE][TAB]`):
 <img src="https://raw.githubusercontent.com/omakoto/compromise/master/img/compromise-adb.png" width=600>
 
## Installing ADB and/or Go Completion

Assuming you have the `go` command installed, just run the following commands in your shell.
```bash
# Get and install the binaries.
go get -v -u github.com/omakoto/compromise/src/cmds/...

# Then to install completion, run them on your shell.
# Add them to your shell's RC file (.bashrc or .zshrc) to make them persistent.
. <(compromise-adb) # Install ADB / fastboot / atest / m* completion
. <(compromise-go)  # Install Go completion
```
 
 *NOTE `go run` won't work; you need to actually compile them.*
 
### Creating Aliases to ADB Subcommands
 - `compromise-adb` also installs completion for some "shorthand" commands,
so if you have following aliases, completion will work for them too.

```bash
alias logcat="adb logcat"
alias dumpsys="adb shell dumpsys"
alias cmd="adb shell cmd"
alias am="adb shell am"
alias pm="adb shell pm"
alias settings="adb shell settings"
alias akill="adb shell akill"
alias akillall="adb shell akillall"
  :
```  
For the full supported command name list, see [the source code](src/cmds/compromise-adb/adb.go).
 
 - If you do not want to install completion for all the listed commands
   in the source file, pass the command name you want to use as arguments. Example: 

```bash
. <(compromise-adb adb dumpsys) # Only install competion for the adb and dumpsys commands.  

# or, disable selectively.

. <(compromise-adb - atest) # Install everything except for the atest completion.  
```

## Known Issues

 - Not heavily tested on Zsh yet.

## TODOs
 - Write tests
 - Add menu completion for bash using [FZF](https://github.com/junegunn/fzf).
 - Cache candidates and only refresh it when "dependency" files are newer. (speed up modulelist)
 
