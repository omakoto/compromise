# Compromise / Compromise-{adb,go}

A Go framework for writing shell completion for Bash / Zsh.

## Bundled completion
 - ADB (Android Debug Bridge)
 - Go


## Caveat
 It's still in a beta stage. Details are subject to change, but feedback is welcome.  

## Installing ADB and/or Go completion

```sh
go get -u github.com/omakoto/compromise/src/cmds/...

. <(compromise-adb) # Install ADB completion
. <(compromise-go)  # Install Go completion
```

## Features

 - Define completion in [a simple-ish language](src/cmds/compromise-go/go.go) that supports both Bash and Zsh.
 - Generate candidates with [custom Go functions](src/cmds/compromise-adb/adb.go).
 - Show description not only on Zsh but on Bash too.

## Known Issues

 - Not heavily tested on Zsh yet.

## TODOs

 - 