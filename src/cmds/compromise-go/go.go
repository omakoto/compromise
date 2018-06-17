package main

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
)

func init() {
	compfunc.Register("takeToolName", takeToolName)
}

func main() {
	compmain.Main(spec)
}

func takeToolName() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithMap("go tool", func(line int, s string) string {
		return s
	})
}

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).Json() + `
@command go

@switch
		build		# compile packages and dependencies
			@call :build
		clean		# remove object files and cached files
			@call :clean
		doc			# show documentation for package or symbol
			@call :doc
		env			# print Go environment information
			@call :env

		//This is not used very often and conflict with "build" so commenting out.
		//bug		  # start a bug report

		fix			# update packages to use new APIs
			@call :fix
		fmt			# gofmt (reformat) package sources
			@call :fmt
		generate	# generate Go files by processing source
			@call :generate
		get			# download and install packages and dependencies
			@call :get
		install		# compile and install packages and dependencies
			@call :install
		list		# list packages
			@call :list
		run			# compile and run Go program
			@call :run
		test		# test packages
			@call :test
		tool		# run specified go tool
			@call :tool
		version		# print Go version

		vet			# report likely mistakes in packages
			@call :vet
		help		# show help for a command or a topic
			@call :help


@label :help
	@switch
		build		# compile packages and dependencies
		clean		# remove object files and cached files
		doc			# show documentation for package or symbol
		env			# print Go environment information
		bug			# start a bug report
		fix			# update packages to use new APIs
		fmt			# gofmt (reformat) package sources
		generate	# generate Go files by processing source
		get			# download and install packages and dependencies
		install		# compile and install packages and dependencies
		list		# list packages
		run			# compile and run Go program
		test		# test packages
		tool		# run specified go tool
		version		# print Go version
		vet			# report likely mistakes in packages

		c			# calling between Go and C
		buildmode	# build modes
		cache		# build and test caching
		filetype	# file types
		gopath		# GOPATH environment variable
		environment # environment variables
		importpath	# import path syntax
		packages	# package lists
		testflag	# testing flags
		testfunc	# testing functions

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
		
		-asmflags					   #\
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
		
		@call :packages

@label :clean
		@switchloop "^-"
				@call :buildflags
		
		@call :packages

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
			@call :package
			@any #[package|[package.]symbol[.methodOrField]]


@label :env
		@switch "^-"
				-json # prints the environment in JSON format instead of as a shell script.
		@loop
			@any # var name

@label :fix
		@call :packages

@label :fmt
		@switchloop "^-"
			-n #prints commands that would be executed.
			-x #prints commands as they are executed.

		@call :packages

@label :generate
		@switchloop "^-"
			-run #\
				# if non-empty, specifies a regular expression to select \
				# directives whose full original source text (excluding \
				# any trailing spaces and final newline) matches the \
				# expression.
				@any # pattern

			-v	# prints the names of packages and files as they are processed.
			-n	# prints commands that would be executed.
			-x	# prints commands as they are executed.

		@call :gofiles

@label :get
		@switchloop "^-"
			-d	# instructs get to stop after downloading the packages; that is, \
				# it instructs get not to install the packages.
			
			-f	# valid only when -u is set, forces get -u not to verify that \
				# each package has been checked out from the source control repository \
				# implied by its import path. This can be useful if the source is a local fork \
				# of the original.
			
			-fix # instructs get to run the fix tool on the downloaded packages \
				# before resolving dependencies or building the code.
			
			-insecure # permits fetching from repositories and resolving \
				# custom domains using insecure schemes such as HTTP. Use with caution.
			
			-t # instructs get to also download the packages required to build \
				# the tests for the specified packages.
			
			-u	# instructs get to use the network to update the named packages \
				# and their dependencies. By default, get uses the network to check out \
				# missing packages but does not use it to look for updates to existing packages.
			
			-v # enables verbose progress and debug output.
			@call :buildflags
	
		@call :packages


@label :install
		@switchloop "^-"
			-i		#installs the dependencies of the named packages as well.
			@call :buildflags

		@call :packages

@label :list
		@switchloop "^-"
			-json	# causes the package data to be printed in JSON format instead of using the template format.
			-e		# changes the handling of erroneous packages, those that cannot be found or are malformed.
			-f		# flag specifies an alternate format for the list, using the syntax of package template.
			@call :buildflags

		@call :packages

@label :run
		@switchloop "^-"
			-exec	# -exec flag is given, 'go run' invokes the binary using xprog: 'xprog a.out arguments...'.
				@any # xprog

			@call :buildflags

		@call :gofiles
	
		//@loop
		//	@any # argument	 // This doesn't work because the above call will eat all arguments...

@label :test
		@switchloop "^-"
			-args #\
				# Pass the remainder of the command line (everything after -args) \
				# to the test binary, uninterpreted and unchanged. \
				# Because this flag consumes the remainder of the command line, \
				# the package list (if present) must appear before this flag.
	
			-c #\
				# Compile the test binary to pkg.test but do not run it \
				# (where pkg is the last element of the package's import path). \
				# The file name can be changed with the -o flag.
	
			-exec #\
				# Run the test binary using xprog. The behavior is the same as \
				# in 'go run'. See 'go help run' for details.
				@any # xprog
	
			-i #\
				# Install packages that are dependencies of the test. \
				# Do not run the test.
	
			-json #\
				# Convert test output to JSON suitable for automated processing. \
				# See 'go doc test2json' for the encoding details.
	
			-o #\
				# Compile the test binary to the named file. \
				# The test still runs (unless -c or -i is specified).
				@cand takeFile

			@call :buildflags

		@call :packages

@label :tool
		@switchloop "^-"
			-n # causes tool to print the command that would be executed but not execute it.

		@cand takeToolName

		@loop
			@cand takeFile

@label :vet
		@switchloop "^-"
			-n # prints commands that would be executed.
			-x # prints commands as they are executed.
			@call :buildflags

		@call :packages

@label :packages
	@loop
		@call :package

@label :package
	@cand takeDir

@label :gofiles
	@loop
		@cand takeFile "\\.go$"

`
