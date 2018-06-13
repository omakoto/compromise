package main

import (
	"bytes"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
	"github.com/omakoto/go-common/src/fileutils"
	"github.com/omakoto/go-common/src/shell"
	"github.com/ungerik/go-dry"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	targetOption = "" // Either "-d" or "-e"
	targetSerial = "" // e.g. "emulator-5554"

	targetUserId           = "" // e.g. "0", "10", "current"
	targetSettingNamespace = "" // e.g. "global"
)

func init() {
	compfunc.Register("takeDevicePackage", takeDevicePackage)
	compfunc.Register("takeDeviceFile", takeDeviceFile)
	compfunc.Register("takeDeviceCommand", takeDeviceCommand)
	compfunc.Register("takeDeviceSerial", takeDeviceSerial)

	compfunc.Register("takeService", takeService)
	compfunc.Register("takeSettingKey", takeSettingKey)
	compfunc.Register("takeUserId", takeUserId)
	compfunc.Register("takePid", takePid)
	compfunc.Register("takeProcessName", takeProcessName)

	compfunc.Register("takeLogcatFilter", takeLogcatFilter)

	compfunc.Register("takeBuildModule", takeBuildModule)

	compfunc.Register("setTargetDevice", compfunc.SetString(&targetOption, "-d"))
	compfunc.Register("setTargetEmulator", compfunc.SetString(&targetOption, "-e"))
	compfunc.Register("setTargetSerial", compfunc.SetLastSeenString(&targetSerial))

	compfunc.Register("setUserId", compfunc.SetLastSeenString(&targetUserId))
	compfunc.Register("setSettingsNamespace", compfunc.SetLastSeenString(&targetSettingNamespace))
}

// Build "adb [-e|-d] [-s serial]"
func adb() string {
	b := bytes.Buffer{}
	b.WriteString("adb")
	if targetOption != "" {
		b.WriteString(" ")
		b.WriteString(targetOption)
	}
	if len(targetSerial) > 0 {
		b.WriteString(" -s ")
		b.WriteString(targetSerial)
	}
	return b.String()
}

// Generate on-device package lists.
func takeDevicePackage() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithMap(adb()+` shell pm list packages 2>/dev/null || true`, func(line int, s string) string {
		return strings.Replace(s, "package:", "", 1)
	})
}

// Generate on-device file lists.
func takeDeviceFile(ctx compromise.CompleteContext) compromise.CandidateList {
	tok := ctx.WordAtCursor(0)
	return compfunc.BuildCandidateListFromCommandWithBuilder(adb()+` shell "ls -pd1 `+shell.Escape(tok)+`* 2>/dev/null || true"`,
		func(line int, s string, b *compromise.CandidateBuilder) {
			b.Value(s).Continues(true) // Continues(true) suppresses a space after a candidate.
		})
}

// Generate on-device command lists.
func takeDeviceCommand(ctx compromise.CompleteContext) compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithBuilder(adb()+` shell 'for n in ${PATH//:/ } ; do ls -1 "$n" ; done 2>/dev/null' || true`,
		func(line int, s string, b *compromise.CandidateBuilder) {
			b.Value(s)
		})
}

// Generate lists of device serials.
func takeDeviceSerial() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithMap("adb devices", func(line int, s string) string {
		if line == 0 {
			return ""
		}
		return compfunc.FieldAt(s, 0, true)
	})
}

// Generate lists of services.
func takeService() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithMap(adb()+" shell dumpsys -l", func(line int, s string) string {
		if line == 0 {
			return ""
		}
		return strings.Trim(s, " \t\r\n")
	})
}

// Generate lists of setting keys.
func takeSettingKey() compromise.CandidateList {
	if targetSettingNamespace == "" {
		targetSettingNamespace = "global" // Default, just in case.
	}
	return compfunc.BuildCandidateListFromCommandWithMap(adb()+" shell settings list "+targetSettingNamespace, func(line int, s string) string {
		if line == 0 {
			return ""
		}
		return compfunc.FieldAtWithSeparator(s, 0, "=", true)
	})
}

// Generate lists of user IDs on the device.
func takeUserId() compromise.CandidateList {
	re := regexp.MustCompile(`UserInfo{(\d+)`)
	return compfunc.BuildCandidateListFromCommandWithMap(adb()+" shell dumpsys user", func(line int, s string) string {
		if m := re.FindStringSubmatch(s); len(m) > 0 {
			return m[1]
		}
		return ""
	})
}

func takePid() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithBuilder(adb()+" shell ps -o PID,NAME", func(line int, s string, builder *compromise.CandidateBuilder) {
		if line == 0 {
			return
		}
		fields := strings.Fields(strings.Trim(s, " \t\r"))
		builder.Value(compfunc.GetField(fields, 0))
		builder.Help(compfunc.GetField(fields, 1))
	})
}

func takeProcessName() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithMap(adb()+" shell ps -oNAME", func(line int, s string) string {
		if line == 0 || strings.HasPrefix(s, "[") {
			return ""
		}
		return s
	})
}

func takeLogcatFilter() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		// We can't really complete a filter, but if a filter ends with :, then we can suggest
		// a priority.
		ret := make([]compromise.Candidate, 0)
		if strings.HasSuffix(prefix, ":") {
			for _, p := range "VDIWEFS" {
				ret = append(ret, compromise.NewCandidateBuilder().Value(prefix+string(p)).Build())
			}
		} else if strings.ContainsRune(prefix, ':') {
			// If the current word already contains a colon, then it's complete.
			ret = append(ret, compromise.NewCandidateBuilder().Value(prefix).Build())
		} else {
			// Otherwise, nothing to suggest but show a help.
			ret = append(ret, compfunc.AnyWithHelp("<log component>[:proirity, any of V D I W E F S]"))
		}

		return ret
	})
}

func takeBuildModuleReal() compromise.CandidateList {
	// This one actually reads as json, but it's a bit slow...
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		var data map[string]interface{}
		err := dry.FileUnmarshallJSON(path.Join(os.Getenv("OUT"), "module-info.json"), &data)
		if err != nil {
			return nil
		}
		ret := make([]compromise.Candidate, 0)
		for k := range data {
			ret = append(ret, compromise.NewCandidateBuilder().Value(k).Build())
		}
		return ret
	})
}

func takeBuildModule(args []string) compromise.CandidateList {
	var re *regexp.Regexp
	if len(args) > 0 {
		re = regexp.MustCompile(args[0])
	}

	// Cheat version. It assumes the content looks like the following (i.e. one module each line):
	/*
		{
		  "1x.sh": { "class": ["EXECUTABLES"],  "path": ["vendor/google_devices/marlin...
		  "26.0.cil": { "class": ["ETC"],  "path": ["system/sepolicy"],  "tags": ["opt...
		  :
		}
	*/
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		file := path.Join(os.Getenv("OUT"), "module-info.json")
		if !fileutils.FileExists(file) {
			return nil
		}
		lines, err := dry.FileGetLines(file)
		if err != nil {
			return nil
		}
		ret := make([]compromise.Candidate, 0)
		for _, line := range lines {
			line = strings.TrimLeft(line, " \t\"")
			p := strings.IndexByte(line, '"')
			if p <= 0 {
				continue
			}
			name := line[:p]
			if re == nil || re.FindStringIndex(name) != nil {
				ret = append(ret, compromise.NewCandidateBuilder().Value(name).Build())
			}
		}
		return ret
	})
}

func main() {
	compmain.Main(spec)
}

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).Json() + `
@command adb
@command fastboot :fastboot
@command atest :atest
@command am :am
@command pm :pm
@command settings :settings
@command asettings :settings
@command cmd :cmd
@command acmd :cmd
@command dumpsys :dumpsys
@command requestsync :requestsync
@command akill :kill
@command akillall :killall
@command logcat :logcat

@switchloop "^-"
	-a # listen on all network interfaces, not just localhost

	-d # use USB device (error if multiple devices connected)
		@go_call setTargetDevice

	-e # use TCP/IP device (error if multiple TCP/IP devices available)
		@go_call setTargetEmulator

	-s # <SERIAL> use device with given serial (overrides $ANDROID_SERIAL)
		@cand TakeDeviceSerial
		@go_call setTargetSerial

	-t # <ID> use device with given transport id
		@any	 # <ID> use device with given transport id

	-H # name of adb server host [default=localhost]

	-P # port of adb server [default=5037]

	-L # <SOCKET> listen on given socket for adb server [default=tcp:localhost:5037]
		@any	# <SOCKET> listen on given socket for adb server [default=tcp:localhost:5037]

@switch
	devices		# list connected devices (-l for long output)
		-l		# long output

	help		# show this help message
	version		# show version num

// networking:
	connect			# HOST[:PORT] connect to a device via TCP/IP [default port=5555]
		@any		# HOST[:PORT] connect to a device via TCP/IP [default port=5555]
	disconnect		# HOST[:PORT]] disconnect from given TCP/IP device [default port=5555], or all
		@any		# HOST[:PORT]] disconnect from given TCP/IP device [default port=5555], or all
	forward			# list all forward socket connections
		@switch "^-"
			--list		# list all forward socket connections
				@finish
			--remove	# LOCAL remove specific forward socket connection
				@any	# LOCAL remove specific forward socket connection
				@finish
			--remove-all # remove all forward socket connections
				@finish
			--no-rebind	# forward socket connection

		@any	# LOCAL: tcp:<port>, localabstract:<domainsocket>, localfilesystem:<domainsocket>,dev:<cdev> 
		@any	# REMOTE: tcp:<port>, localabstract:<domainsocket>, localfilesystem:<domainsocket>,dev:<cdev>,jdwp:<pid>

	ppp			# run PPP over USB
		@any	# PPP

	reverse
		@switch
			--list		# list all reverse socket connections from device
				@finish
			--remove	# remove specific reverse socket connection
				@any	# REMOTE: tcp:<port>, localabstract:<domainsocket>, localfilesystem:<domainsocket>,dev:<cdev>,jdwp:<pid>
				@finish
			--remove-all # remove all reverse socket connections from device
				@finish
			--no-rebind	# reverse socket connection
				@finish

		@any	# REMOTE: tcp:<port>, localabstract:<domainsocket>, localfilesystem:<domainsocket>,dev:<cdev>,jdwp:<pid>
		@any	# LOCAL: tcp:<port>, localabstract:<domainsocket>, localfilesystem:<domainsocket>,dev:<cdev> 

// file transfer:
	push		# copy local files/directories to device
		@switch "^-"
			--sync		# only push files that are newer on the host than the device
		@cand takeFile
		@cand takeDeviceFile

	pull		# copy files/dirs from device
		@switch "^-"
			-a			# preserve file timestamp and mode
		@cand takeDeviceFile
		@cand takeFile

	sync		# sync a local build from $ANDROID_PRODUCT_OUT to the device (default all)
		@switch "^-"
			-l	# list but don't copy
		@switch
			system
			vendor
			oem
			data
			all

// app installation:
	install				# push package(s) to the device and install them
		@call :install-options
		@cand takeFile ".*\\.apk"

	install-multiple	# push packages to the device and install them
		@call :install-options
		@loop
			@cand takeFile ".*\\.apk"

	uninstall			# remove this app package from the device
		@switch "^-"
			-k			# keep the data and cache directories
		@cand takeDevicePackage

// backup/restore:
	backup
		@call :bu-backup
	restore
		@call :bu-restore

// debugging:
	bugreport	# write bugreport to given PATH [default=bugreport.zip];
		@cand takeFile

	jdwp		# list pids of processes hosting a JDWP transport

	logcat		# show device log
		@call :logcat

// security:
	disable-verity		# disable dm-verity checking on userdebug builds
	enable-verity		# re-enable dm-verity checking on userdebug builds
	keygen				# generate adb public/private key; private key stored in FILE, public key stored in FILE.pub (existing files overwritten)
		@cand takeFile

// scripting:
	wait-for-device			# wait for device to be in the given state
	wait-for-recovery		# wait for device to be in the given state
	wait-for-sideload		# wait for device to be in the given state
	wait-for-bootloader		# wait for device to be in the given state

	get-state				# print offline | bootloader | device
	get-serialno			# print <serial-number>
	get-devpath				# print <device-path>

	remount					# remount /system, /vendor, and /oem partitions read-write

	reboot					# reboot the device; defaults to booting system image but supports bootloader and recovery too
		@switch
			bootloader
			recovery
			sideload
			sideload-auto-reboot
	reboot-bootloader		# reboot the device into boot loader

	sideload				# sideload the given full OTA package
		@cand takeFile

	root					# restart adbd with root permissions
	unroot					# restart adbd without root permissions
	usb					# restart adb server listening on USB
	tcpip					# restart adb server listening on TCP on PORT
		@any				# PORT restart adb server listening on TCP on PORT

// internal debugging:
	start-server			# ensure that there is a server running
	kill-server				# kill the server if it is running
	reconnect				# kick connection from host side to force reconnect
		@switch
			device		   # kick connection from device side to force reconnect
			offline		   # reset offline/unauthorized devices to force reconnect

	shell		# run remote shell command (interactive shell if no command given)
		@switchloop "^-"
			-e		# <CHAR> choose escape character, or "none"; default '~'
				// Can't add "none" as a @switch-candidate, because it'll always be selected.
				@any # <CHAR> escape character, or "none"

			-n		# don't read from stdin
			-T		# disable PTY allocation
			-t		# force PTY allocation
			-x		# disable remote exit codes and stdout/stderr separation
		@switch
			dumpsys	# Dump system service
				@call :dumpsys
	
			cmd		# Execute a aystem server command
				@call :cmd
			am		# Activity manager command
				@call :am
			pm		# Package manager command
				@call :pm
			settings	# SettingsProvider command
				@call :settings
			logcat	# show device log
				@call :logcat

			requestsync	# SyncManager command
				@call :requestsync

			kill		# Kill process by PID
				@call :kill
			killall		# Kill process by name
				@call :killall
			@cand takeDeviceCommand

@finish

@label :install-options
	@switchloop "^-"
		-l		# forward lock application
		-r		# replace existing application
		-t		# allow test packages
		-s		# install application on sdcard
		-d		# allow version code downgrade (debuggable packages only)
		-p		# partial application install (install-multiple only)
		-g		# grant all runtime permissions

@label :am
	@switch
		start-activity			# Start an Activity.
			@switchloop "^-"
				@call :intent_flags

				-D # enable debugging
				-N # enable native debugging
				-W # wait for launch to complete
				--start-profiler # start profiler and send results to <FILE>
					@cand takeFile
				--sampling # use sample profiling with INTERVAL microseconds between samples (use with --start-profiler)
					@any # INTERVAL microseconds between samples
				--streaming # stream the profiling output to the specified file (use with --start-profiler)
				-P # like above, but profiling stops when app goes idle
					@cand takeFile
				--attach-agent # attach the given agent before binding
					@any # agent
				-R # repeat the activity launch <COUNT> times.	Prior to each repeat, the top activity will be finished.
					@any # COUNT
				-S # force stop the target app before starting the activity
				--track-allocation # enable tracking of object allocations
				--stack # Specify into which stack should the activity be put.
					@any  #STACK_ID
				@call :take_user_id
			
			@call :intent_body

	// TODO Implement other commands...
		
@label :pm
	@switch
		dump	# dump package
			@cand takeDevicePackage

	// TODO Implement other commands...

@label :dumpsys
	@cand takeService

@label :cmd
	@cand takeService

@label :settings
	@switch
		//	 get [--user <USER_ID> | current] NAMESPACE KEY
		get			# Retrieve the current value of KEY
			@call :take_user_id
			@call :settings_namespace	
			@cand takeSettingKey
		
		//	 put [--user <USER_ID> | current] NAMESPACE KEY VALUE [TAG] [default]
		put			# Change the contents of KEY to VALUE
			@call :take_user_id
			@call :settings_namespace
			@cand takeSettingKey
			@any	# <value> value to set
			@any	# <tag> 
			@switch
				default # {default} to set as the default, case-insensitive only for global/secure namespace
	
		//	 delete NAMESPACE KEY
		delete		# Delete the entry for KEY
			@call :settings_namespace
			@cand takeSettingKey
		
		//	 reset [--user <USER_ID> | current] NAMESPACE {PACKAGE_NAME | RESET_MODE}
		reset		# Reset the global/secure table for a package with mode
			@call :take_user_id
			@call :settings_namespace
			@switch
				@cand takeDevicePackage
				untrusted_defaults
				untrusted_clear
				trusted_defaults
	
		//	 list NAMESPACE
		list	# Print all defined keys
			@call :settings_namespace
		
// settings " --user [ X | current ] "
@label :take_user_id	
	@switch "^-"
		--user
			@switch
				@cand takeUserId
					@go_call setUserId
				current
					@go_call setUserId 

// settings global put " [ global | system | secure ] "
@label :settings_namespace
	@switch
		global
			@go_call setSettingsNamespace 
		system
			@go_call setSettingsNamespace 
		secure
			@go_call setSettingsNamespace 

@label :requestsync


@label :bu-backup
@label :bu-restore


@label :kill
	@switchloop "^-"
		-s	# specify signal
			@switch
				HUP # Hangup
				INT # Interrupt
				QUIT # Quit
				ILL # Illegal instruction
				TRAP # Trap
				ABRT # Aborted
				BUS # Bus error
				FPE # Floating point exception
				KILL # Killed
				USR1 # User signal 1
				SEGV # Segmentation fault
				USR2 # User signal 2
				PIPE # Broken pipe
				ALRM # Alarm clock
				TERM # Terminated
				STKFLT # Stack fault
				CHLD # Child exited
				CONT # Continue
				STOP # Stopped (signal)
				TSTP # Stopped
				TTIN # Stopped (tty input)
				TTOU # Stopped (tty output)
				URG # Urgent I/O condition
				XCPU # CPU time limit exceeded
				XFSZ # File size limit exceeded
				VTALRM # Virtual timer expired
				PROF # Profiling timer expired
				WINCH # Window size changed
				IO # I/O possible
				PWR # Power failure
				SYS # Bad system call
		-l	# list signals
	@loop
		@cand takePid

@label :killall
	@switchloop "^-"
		-i		# ask for confirmation before killing
		-l		# print list of all available signals
		-q		# don't print any warnings or error messages
		-s		# send SIGNAL instead of SIGTERM
		-v		# report if the signal was successfully sent
	@loop
		@cand takeProcessName

@label :logcat
	@switchloop "^-"
			--help				# Show help
			-s					# Set default filter to silent. Equivalent to filterspec '*:S'
			-f|--file			# Log to file. Default is stdout
				@cand takeFile // TODO It'd be great if we can show help for it too.
			-r|--rotate-kbytes	# Rotate log every kbytes. Requires -f option
				@any			# <kbytes> Rotate log every kbytes. Requires -f option
			-n|--rotate-count	# Sets max number of rotated logs to <count>, default 4
				@any			# <count> Sets max number of rotated logs to <count>, default 4
			--id				# If the signature id for logging to file changes, then clear the fileset and continue
				@any			# <id>
			-v					# Sets log print format verb and adverbs
				@switch
					brief
					help
					long
					process
					raw
					tag
					thread
					threadtime
					time
					uid
			-D|--dividers		# Print dividers between each log buffer
			-c|--clear			# Clear (flush) the entire log and exit
			-d					# Dump the log and then exit (don't block)
			-e|--regex			# Only print lines where the log message matches <expr> where <expr> is a regular expression
				@any			# <expr> Only print lines where the log message matches <expr> where <expr> is a regular expression
			-m|--max-count		# Quit after printing <count> lines
				@any			# <count> Quit after printing <count> lines
			--print				# Paired with --regex and --max-count to let content bypass regex filter but still stop at number of matches.
			-t					# Print only the most recent lines (implies -d)
				@any			# <count> or '<time>' 
			-T					# Print only the most recent lines (does not implies -d)
				@any			# <count> or '<time>' 
			-g|--buffer-size	# Get the size of the ring buffer
			-G|--buffer-size	# Set size of log ring buffer, may suffix with K or M.
				@any			# <size> Set size of log ring buffer, may suffix with K or M.
			-L|--last			# Dump logs from prior to last reboot
			-b|--buffer			# Request alternate ring buffer
				@switch
					main
					system
					radio
					events
					crash
					default
					all
			-d					# Dump the log and then exit (don't block)
			-B|--binary			# Output the log in binary
			-S|--statistics		# Output statistics
			-p|--prune			# Print prune white and ~black list
			--pid				# Only prints logs from the given pid
				@cand takePid
			--wrap				# Sleep for 2 hours or when buffer about to wrap whichever comes first


	@loop
		@cand takeLogcatFilter

@label :intent
	@switchloop "^-"
		@call :intent_flags
	@call :intent_body

@label :intent_flags
		-a #<ACTION>
			@any #ACTION
		-d #<DATA_URI>
			@any #DATA_URI
		-t #<MIME_TYPE>
			@any #MIME_TYPE
		-c #<CATEGORY>
			@any #CATEGORY
		-e|--es #<EXTRA_KEY> <EXTRA_STRING_VALUE>
			@any #EXTRA_KEY
			@any #EXTRA_STRING_VALUE
		--esn #<EXTRA_KEY> ...
			@any #EXTRA_KEY
		--ez #<EXTRA_KEY> <EXTRA_BOOLEAN_VALUE>
			@any #EXTRA_KEY
			@any #EXTRA_BOOLEAN_VALUE
		--ei #<EXTRA_KEY> <EXTRA_INT_VALUE>
			@any #EXTRA_KEY
			@any #EXTRA_INT_VALUE
		--el #<EXTRA_KEY> <EXTRA_LONG_VALUE>
			@any #EXTRA_KEY
			@any #EXTRA_LONG_VALUE
		--ef #<EXTRA_KEY> <EXTRA_FLOAT_VALUE>
			@any #EXTRA_KEY
			@any #EXTRA_FLOAT_VALUE
		--eu #<EXTRA_KEY> <EXTRA_URI_VALUE>
			@any #EXTRA_KEY
			@any #EXTRA_URI_VALUE
		--ecn #<EXTRA_KEY> <EXTRA_COMPONENT_NAME_VALUE>
			@any #EXTRA_KEY
			@cand takeDevicePackage // TODO Change to ComponentName
	
		--eia #<EXTRA_KEY> <EXTRA_INT_VALUE>[,<EXTRA_INT_VALUE...] \
			#(mutiple extras passed as Integer[])
			@any #EXTRA_KEY
			@any #EXTRA_INT_VALUE,EXTRA_INT_VALUE,...
		--eial #<EXTRA_KEY> <EXTRA_INT_VALUE>[,<EXTRA_INT_VALUE...] \
			#(mutiple extras passed as List<Integer>)
			@any #EXTRA_KEY
			@any #EXTRA_INT_VALUE,EXTRA_INT_VALUE,...
		--ela #<EXTRA_KEY> <EXTRA_LONG_VALUE>[,<EXTRA_LONG_VALUE...] \
			#(mutiple extras passed as Long[])
			@any #EXTRA_KEY
			@any #EXTRA_LONG_VALUE,EXTRA_LONG_VALUE,...
		--elal #<EXTRA_KEY> <EXTRA_LONG_VALUE>[,<EXTRA_LONG_VALUE...] \
			#(mutiple extras passed as List<Long>)
			@any #EXTRA_KEY
			@any #EXTRA_LONG_VALUE,EXTRA_LONG_VALUE,...
		--efa #<EXTRA_KEY> <EXTRA_FLOAT_VALUE>[,<EXTRA_FLOAT_VALUE...] \
			#(mutiple extras passed as Float[])
			@any #EXTRA_KEY
			@any #EXTRA_FLOAT_VALUE,EXTRA_FLOAT_VALUE,...
		--efal #<EXTRA_KEY> <EXTRA_FLOAT_VALUE>[,<EXTRA_FLOAT_VALUE...] \
			#(mutiple extras passed as List<Float>)
			@any #EXTRA_KEY
			@any #EXTRA_FLOAT_VALUE,EXTRA_FLOAT_VALUE,...
		--esa #<EXTRA_KEY> <EXTRA_STRING_VALUE>[,<EXTRA_STRING_VALUE...] \
			#(mutiple extras passed as String[]; to embed a comma into a string, \
			#escape it using "\,")
			@any #EXTRA_KEY
			@any #EXTRA_STRING_VALUE,EXTRA_STRING_VALUE,...
		--esal #<EXTRA_KEY> <EXTRA_STRING_VALUE>[,<EXTRA_STRING_VALUE...] \
			#(mutiple extras passed as List<String>; to embed a comma into a string, \
			#escape it using "\,")
			@any #EXTRA_KEY
			@any #EXTRA_STRING_VALUE,EXTRA_STRING_VALUE,...

		-f # <FLAG>
			@any # Intent flags (0xHEX, 0OCT or decimal) 
		--grant-read-uri-permission
		--grant-write-uri-permission
		--grant-persistable-uri-permission
		--grant-prefix-uri-permission
		--debug-log-resolution
		--exclude-stopped-packages
		--include-stopped-packages
		--activity-brought-to-front
		--activity-clear-top
		--activity-clear-when-task-reset
		--activity-exclude-from-recents
		--activity-launched-from-history
		--activity-multiple-task
		--activity-no-animation
		--activity-no-history
		--activity-no-user-action
		--activity-previous-is-top
		--activity-reorder-to-front
		--activity-reset-task-if-needed
		--activity-single-top
		--activity-clear-task
		--activity-task-on-home
		--receiver-registered-only
		--receiver-replace-pending
		--receiver-foreground
		--receiver-no-abort
		--receiver-include-background
		--selector

@label :intent_body
	@switch
		// TODO A URI is accepted here.	 
		@cand takeDevicePackage // TODO Or, a component name. Need to add a pm command to get components.  











@label :fastboot // TODO support device serial completion.
	@switchloop "^-"
		-w										# Erase userdata and cache (and format \
												# if supported by partition type).
		-u										# Do not erase partition before \
												# formatting.
		-s										# Specify a device. For USB, provide either \
												# a serial number or path to device port. \
												# For ethernet, provide an address in the \
												# form <protocol>:<hostname>[:port] where \
												# <protocol> is either tcp or udp.
			@any								# <serial>
		-c										# Override kernel commandline.
			@any								# <commandline>
			
		-i										# Specify a custom USB vendor id.
			@any								# <vendor id>
		-b|--base								# Specify a custom kernel base \
												# address (default: 0x10000000).
			@any								# <base addr>
		--kernel-offset							# Specify a custom kernel offset. \
												# (default: 0x00008000)
			@any								# <base addr>
		--ramdisk-offset						# Specify a custom ramdisk offset. \
												# (default: 0x01000000)
			@any								# <base addr>
		--tags-offset							# Specify a custom tags offset. \
												# (default: 0x00000100)
			@any								# <base addr>
		-n|--page-size							# Specify the nand page size \
												# (default: 2048).
			@any								# <page size>
		-S										# Automatically sparse files greater \
												# than 'size'. 0 to disable.
			@any								# <size>[K|M|G]	 
		--slot									# Specify slot name to be used if the \
												# device supports slots. All operations \
												# on partitions that support slots will \
												# be done on the slot specified.
			@any								# <slot>  
		-a|--set-active							# Sets the active slot.
			@any								# <slot>  
		--skip-secondary						# Will not flash secondary slots when \
												# performing a flashall or update. This \
												# will preserve data on other slots.
		--skip-reboot							# Will not reboot the device when \
												# performing commands that normally \
												# trigger a reboot.
		--disable-verity						# Set the disable-verity flag in the \
												# the vbmeta image being flashed.
		--disable-verification					# Set the disable-verification flag in \
												# the vbmeta image being flashed.
		--wipe-and-use-fbe						# On devices which support it, \
												# erase userdata and cache, and \
												# enable file-based encryption
		--unbuffered							# Do not buffer input or output.
		--version								# Display version.
		--header-version						# Set boot image header version while \
												# using flash:raw and boot commands to \
												# to create a boot image.
		-h|--help								# show this message.


	@switch
		update									# Reflash device from update.zip. \
												# Sets the flashed slot as active.
			@cand takeFile
		flashall								# Flash boot, system, vendor, and -- \
												# if found -- recovery. If the device \
												# supports slots, the slot that has \
												# been flashed to is set as active. \
												# Secondary images may be flashed to \
												# an inactive slot.
		flash									# Write a file to a flash partition.
			@any								# Partition name
			@cand takeFile
		flashing
			@switch
				 lock							# Locks the device. Prevents flashing.
				 unlock							# Unlocks the device. Allows flashing \
												# any partition except \
												# bootloader-related partitions.
				 lock_critical					# Prevents flashing bootloader-related \
												# partitions.
				 unlock_critical				# Enables flashing bootloader-related \
												# partitions.
				 get_unlock_ability				# Queries bootloader to see if the \
												# device is unlocked.
				 get_unlock_bootloader_nonce	# Queries the bootloader to get the \
												# unlock nonce.
				 unlock_bootloader				# Issue unlock bootloader using request.
					@any						# <request>
				 lock_bootloader				# Locks the bootloader to prevent \
												# bootloader version rollback.
		erase									# Erase a flash partition.
			@any								# <partition>  
		format									# Format a flash partition. Can \
												# override the fs type and/or size \
												# the bootloader reports. 
			@any								# <partition>  
		getvar									# Display a bootloader variable.
			@any								# <variable>  
		set_active								# Sets the active slot. If slots are \
												# not supported, this does nothing.
			@any								# <slot>  
		boot									# Download and boot kernel.
			@any								# <kernel>	
			@any								# <ramdisk>	 
			@any								# <second>	
		"flash:raw"								# Create bootimage and flash it.
			@any								# <bootable-partition>	
			@any								# <kernel>	
			@any								# <ramdisk>	 
			@any								# <second>	
		devices									# List all connected devices.
			-l									# List all connected devices with device paths.
		continue								# Continue with autoboot.
		reboot									# Reboot device [into bootloader or emergency mode].
			@switch
				bootloader
				emergency
		reboot-bootloader						# Reboot device into bootloader.
		oem										# Executes oem specific command.
			@loop
				@any							# <parameter>
		stage									# Sends contents of <infile> to stage for \
												# the next command. Supported only on \
												# Android Things devices.
			@cand takeFile	  
		get_staged								# Receives data to <outfile> staged by the \
												# last command. Supported only on Android  \
												# Things devices.
			@cand takeFile	  
		help									# Show this help message.


@label :atest
	@switchloop "^-"
		-h|--help			 # Show this help message and exit
		-b|--build			 # Run a build.
		-i|--install		 # Install an APK.
		-t|--test			 # Run the tests. WARNING: Many test configs force cleanup of device after test run. In this case, -d must be used in previous test run to disable cleanup, for -t to work. Otherwise, device will need to be setup again with -i.		--help		# show help

		-s|--serial					# The device to run the test on.
			@cand takeDeviceSerial
		-d|--disable-teardown		# Disables test teardown and cleanup.
		-m|--rebuild-module-info	# Forces a rebuild of the module-info.json file. This may be necessary following a repo sync or when writing a new test.
		-w|--wait-for-debugger		# Only for instrumentation tests. Waits for debugger prior to execution.
		-v|--verbose				# Display DEBUG level logging.
		--generate-baseline			# Generate baseline metrics, run 5 iterations by default. Provide an int argument to specify # iterations.
			@any # Number of iterations.
		
		--generate-new-metrics		# Generate new metrics, run 5 iterations by default. Provide an int argument to specify # iterations. 
			@any # Number of iterations.
						
		--detect-regression			# Run regression detection algorithm. Supply path to baseline and/or new metrics folders.
			@loop
				@cand takeFile
		--
			@break		// TODO Compromise doesn't recoginze it and still suggests flags after --.

	@switchloop
		@cand takeFile	// TODO Support #method1,method2,...
					
		@cand takeBuildModule "(Test|^Bug)"


`
