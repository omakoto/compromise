package main

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compmain"
)

func init() {
}

func main() {
	compmain.Main(spec)
}

// TODO Finish it...

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).Json() + `
@command go

@switch
        build       # compile packages and dependencies
			@call :build
        clean       # remove object files and cached files
			@call :clean
        doc         # show documentation for package or symbol
			@call :doc
        env         # print Go environment information
			@call :env

		//TODO: this is not used very often and conflict with "build" so
		//commenting out for now.
		//But we can revive it once "hidden" candidates are implemented.
        //bug         # start a bug report
		//	@call :bug

        fix         # update packages to use new APIs
			@call :fix
        fmt         # gofmt (reformat) package sources
			@call :fmt
        generate    # generate Go files by processing source
			@call :generate
        get         # download and install packages and dependencies
			@call :get
        install     # compile and install packages and dependencies
			@call :install
        list        # list packages
			@call :list
        run         # compile and run Go program
			@call :run
        test        # test packages
			@call :test
        tool        # run specified go tool
			@call :tool
        version     # print Go version
			@call :version
        vet         # report likely mistakes in packages
			@call :vet
		help        # show help for a command or a topic
			@call :help


@label :help
@switch
        build       # compile packages and dependencies
        clean       # remove object files and cached files
        doc         # show documentation for package or symbol
        env         # print Go environment information
        bug         # start a bug report
        fix         # update packages to use new APIs
        fmt         # gofmt (reformat) package sources
        generate    # generate Go files by processing source
        get         # download and install packages and dependencies
        install     # compile and install packages and dependencies
        list        # list packages
        run         # compile and run Go program
        test        # test packages
        tool        # run specified go tool
        version     # print Go version
        vet         # report likely mistakes in packages

        c           # calling between Go and C
        buildmode   # build modes
        cache       # build and test caching
        filetype    # file types
        gopath      # GOPATH environment variable
        environment # environment variables
        importpath  # import path syntax
        packages    # package lists
        testflag    # testing flags
        testfunc    # testing functions

@label :buildflags
-a #\
        #force rebuilding of packages that are already up-to-date.
-n #\
        #print the commands but do not run them.
-p #\
        #the number of programs, such as build commands or \
        #test binaries, that can be run in parallel. \
        #The default is the number of CPUs available.
			@any #<num parallel>
-race #\
        #enable data race detection. \
        #Supported only on linux/amd64, freebsd/amd64, darwin/amd64 and windows/amd64.
-msan #\
        #enable interoperation with memory sanitizer. \
        #Supported only on linux/amd64, \
        #and only with Clang/LLVM as the host C compiler.
-v #\
        #print the names of packages as they are compiled.
-work #\
        #print the name of the temporary work directory and \
        #do not delete it when exiting.
-x #\
        #print the commands.

-asmflags                      #\
        #arguments to pass on each go tool asm invocation.
			@any #'[pattern=]arg list'

-buildmode=archive #\
        #Build the listed non-main packages into .a files. Packages named \
        #main are ignored.

-buildmode=c-archive #\
        #Build the listed main package, plus all packages it imports, \
        #into a C archive file. The only callable symbols will be those \
        #functions exported using a cgo //export comment. Requires \
        #exactly one main package to be listed.

-buildmode=c-shared #\
        #Build the listed main package, plus all packages it imports, \
        #into a C shared library. The only callable symbols will \
        #be those functions exported using a cgo //export comment. \
        #Requires exactly one main package to be listed.

-buildmode=default #\
        #Listed main packages are built into executables and listed \
        #non-main packages are built into .a files (the default \
        #behavior).

-buildmode=shared #\
        #Combine all the listed non-main packages into a single shared \
        #library that will be used when building with the -linkshared \
        #option. Packages named main are ignored.

-buildmode=exe #\
        #Build the listed main packages and everything they import into \
        #executables. Packages not named main are ignored.

-buildmode=pie #\
        #Build the listed main packages and everything they import into \
        #position independent executables (PIE). Packages not named \
        #main are ignored.

-buildmode=plugin #\
        #Build the listed main packages, plus all packages that they \
        #import, into a Go plugin. Packages not named main are ignored.

-compiler #\
        #name of compiler to use, as in runtime.Compiler (gccgo or gc).
			@switch
				gccgo
				gc

-gccgoflags #\
        #arguments to pass on each gccgo compiler/linker invocation.
			@any #'[pattern=]arg list'

-gcflags #\
        #arguments to pass on each go tool compile invocation.
			@any #'[pattern=]arg list'

-installsuffix #\
        #a suffix to use in the name of the package installation directory, \
        #in order to keep output separate from default builds. \
        #If using the -race flag, the install suffix is automatically set to race \
        #or, if set explicitly, has _race appended to it. Likewise for the -msan \
        #flag. Using a -buildmode option that requires non-default compile flags \
        #has a similar effect.
			@any #<suffix>
-ldflags #\
        #arguments to pass on each go tool link invocation.
			@any #'[pattern=]arg list'

-linkshared #\
        #link against shared libraries previously created with \
        #-buildmode=shared.

-pkgdir #\
        #install and load all packages from dir instead of the usual locations. \
        #For example, when building with a non-standard configuration, \
        #use -pkgdir to keep generated packages in a separate location.
			@cand takeDir

-tags #\
        #a space-separated list of build tags to consider satisfied during the \
        #build. For more information about build tags, see the description of \
        #build constraints in the documentation for the go/build package.
			@any #'tag list' 

-toolexec  #\
        #a program to use to invoke toolchain programs like vet and asm. \
        #For example, instead of running asm, the go command will run \
        #'cmd args /path/to/asm <arguments for asm>'.
			@any #'cmd args'

@label :build
@switchloop "^-"
	@call :buildflags

@cand takedir

@label :clean
@switchloop "^-"
		@call :buildflags

@loop
	@cand takedir


@label :doc
@switchloop "^-"
        -c #\
                #Respect case when matching symbols.
        -cmd #\
                #Treat a command (package main) like a regular package. \
                #Otherwise package main's exported symbols are hidden \
                #when showing the package's top-level documentation.
        -u #\
                #Show documentation for unexported as well as exported \
                #symbols, methods, and fields.
@switchloop
	@cand takedir
	@any #[package|[package.]symbol[.methodOrField]]


@label :env
@label :bug
@label :fix

@label :fmt
@switchloop "^-"
		-n #prints commands that would be executed.
		-x #prints commands as they are executed.

@loop
	@cand takedir

@label :generate
@label :get

@label :install
@switchloop "^-"
		@call :buildflags
		-i 		#installs the dependencies of the named packages as well.
@loop
	@cand takedir

@label :list
@label :run
@label :test
@label :tool
@label :version
@label :vet


`
