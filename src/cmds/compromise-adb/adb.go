package main

import (
	"bytes"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compdebug"
	"github.com/omakoto/compromise/src/compromise/compfunc"
	"github.com/omakoto/compromise/src/compromise/compmain"
	"github.com/omakoto/go-common/src/fileutils"
	"github.com/omakoto/go-common/src/shell"
	"github.com/ungerik/go-dry"
	"io/ioutil"
)

var (
	targetOption = "" // Either "-d" or "-e"
	targetSerial = "" // e.g. "emulator-5554"

	targetUserID           = "" // e.g. "0", "10", "current"
	targetSettingNamespace = "" // e.g. "global"
)

func init() {
	compfunc.Register("takeDevicePackage", takeDevicePackage)
	compfunc.Register("takeDevicePackageComponent", takeDevicePackageComponent)
	compfunc.Register("takePermission", takePermission)
	compfunc.Register("takeDeviceActivity", takeDeviceActivity)
	compfunc.Register("takeDeviceService", takeDeviceService)
	compfunc.Register("takeDeviceReceiver", takeDeviceReceiver)
	compfunc.Register("takeDeviceProvider", takeDeviceProvider)
	compfunc.Register("takeDeviceInstrumentation", takeDeviceInstrumentation)
	compfunc.Register("takeDeviceFile", takeDeviceFile)
	compfunc.Register("takeDeviceCommand", takeDeviceCommand)
	compfunc.Register("takeDeviceSerial", takeDeviceSerial)

	compfunc.Register("takeService", takeService)
	compfunc.Register("takeSettingKey", takeSettingKey)
	compfunc.Register("takeUserID", takeUserID)
	compfunc.Register("takePid", takePid)
	compfunc.Register("takeProcessName", takeProcessName)

	compfunc.Register("takeLogcatFilter", takeLogcatFilter)

	compfunc.Register("takeBuildModule", takeBuildModule)

	compfunc.Register("takeJavaFileMethod", takeJavaFileMethod)

	compfunc.Register("setTargetDevice", compfunc.SetString(&targetOption, "-d"))
	compfunc.Register("setTargetEmulator", compfunc.SetString(&targetOption, "-e"))
	compfunc.Register("setTargetSerial", compfunc.SetLastSeenString(&targetSerial))

	compfunc.Register("setUserId", compfunc.SetLastSeenString(&targetUserID))
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

// Generate on-device permission lists.
func takePermission() compromise.CandidateList {
	return compfunc.BuildCandidateListFromCommandWithMap(adb()+` shell pm list permissions 2>/dev/null || true`, func(line int, s string) string {
		p := "permission:"
		if strings.HasPrefix(s, p) {
			return s[len(p):]
		}
		return ""
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
func takeUserID() compromise.CandidateList {
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

func getPackageComponents(pkg string) (activities, services, receivers, providers, instrumentations, all []string) {
	bdump, err := compfunc.ExecAndGetStdout(adb() + " shell dumpsys package --all-components " + pkg)
	if err != nil {
		return
	}

	var target *[]string

	const indent = "    "
	header := regexp.MustCompile(`^` + indent + `(activities:|services:|receivers:|providers:|instrumentations:|[^ ].*)$`)
	for _, line := range strings.Split(string(bdump), "\n") {
		if header.MatchString(line) {
			// fmt.Fprintf(os.Stderr, "  * %s\n", line)
			target = nil
			switch line {
			case indent + "activities:":
				target = &activities
			case indent + "services:":
				target = &services
			case indent + "receivers:":
				target = &receivers
			case indent + "providers:":
				target = &providers
			case indent + "instrumentations:":
				target = &instrumentations
			}
			if target != nil {
				*target = make([]string, 0)
			}
		}
		if target != nil {
			*target = append(*target, strings.Trim(line, compfunc.Space))
		}
	}
	sort.Strings(activities)
	sort.Strings(services)
	sort.Strings(receivers)
	sort.Strings(providers)
	sort.Strings(instrumentations)

	all = make([]string, 0)
	all = append(all, activities...)
	all = append(all, services...)
	all = append(all, receivers...)
	all = append(all, providers...)
	all = append(all, instrumentations...)
	sort.Strings(all)

	return
}

func getPackageActivities(pkg string) []string {
	ret, _, _, _, _, _ := getPackageComponents(pkg)
	return ret
}

func getPackageServices(pkg string) []string {
	_, ret, _, _, _, _ := getPackageComponents(pkg)
	return ret
}

func getPackageReceivers(pkg string) []string {
	_, _, ret, _, _, _ := getPackageComponents(pkg)
	return ret
}

func getPackageProviders(pkg string) []string {
	_, _, _, ret, _, _ := getPackageComponents(pkg)
	return ret
}

func getPackageInstrumentations(pkg string) []string {
	_, _, _, _, ret, _ := getPackageComponents(pkg)
	return ret
}

func getPackageAllComponents(pkg string) []string {
	_, _, _, _, _, ret := getPackageComponents(pkg)
	return ret
}

func takeDeviceActivity() compromise.CandidateList {
	return takeDeviceComponentInner(getPackageActivities)
}

func takeDeviceService() compromise.CandidateList {
	return takeDeviceComponentInner(getPackageServices)
}

func takeDeviceReceiver() compromise.CandidateList {
	return takeDeviceComponentInner(getPackageReceivers)
}

func takeDeviceProvider() compromise.CandidateList {
	return takeDeviceComponentInner(getPackageProviders)
}

func takeDeviceInstrumentation() compromise.CandidateList {
	return takeDeviceComponentInner(getPackageInstrumentations)
}

func takeDevicePackageComponent() compromise.CandidateList {
	return takeDeviceComponentInner(getPackageAllComponents)
}

func takeDeviceComponentInner(fetcher func(string) []string) compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		p := strings.Index(prefix, "/")
		if p < 0 {
			return takeDevicePackage().GetCandidate(prefix)
		} else if p == 0 {
			return nil
		}
		return compfunc.StringsToCandidates(fetcher(prefix[0:p]), func(line int, s string, b *compromise.CandidateBuilder) {
			b.Value(s)
		})
	})
}

// Extract test-method-looking words from a file.
func findJavaTestMethods(file string) []string {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	// Just extract all "public void" methods.
	pat := regexp.MustCompile(`^\s*public\s+void\s+(\w+)\(`)

	ret := make([]string, 0)
	for _, l := range bytes.Split(b, []byte("\n")) {
		if m := pat.FindSubmatch(l); m != nil {
			ret = append(ret, string(m[1]))
		}
	}
	return ret
}

// Completion for atest-style "Filename#method1,method2,..." arguments.
func takeJavaFileMethod() compromise.CandidateList {
	return compromise.LazyCandidates(func(prefix string) []compromise.Candidate {
		compdebug.Debugf("takeJavaFileMethod prefix=%s\n", prefix)
		sharp := strings.Index(prefix, "#")
		if sharp <= 0 {
			if fileutils.FileExists(prefix) {
				// Argument is a filename. Return [filename] + "#".
				return compromise.StrictCandidates(compromise.NewCandidateBuilder().Value(prefix + "#").Force(true).Continues(true).Build()).GetCandidate("")
			}
			// Doesn't contain a "#", so just do a file completion, but don't append " " after a filename.
			return compfunc.TakeFileWithMapper(`\.java$`, func(b *compromise.CandidateBuilder) {
				b.Continues(true)
			}).GetCandidate(prefix)
		}

		file := prefix[0:sharp]

		resultPrefix := ""

		lastComma := strings.LastIndex(prefix, ",")
		if lastComma > 0 {
			// Everything before the last , will be the prefix.
			resultPrefix = prefix[0:lastComma] + ","
		} else {
			// "," not found, so this will be the prefix.
			resultPrefix = file + "#"
		}

		ret := make([]compromise.Candidate, 0)
		for _, method := range findJavaTestMethods(file) {
			compdebug.Debugf("prefix=%s method=%s\n", resultPrefix, method)
			// Append method names to the result prefix (which is either "filename#" or "filename#method1,method2,")
			ret = append(ret, compromise.NewCandidateBuilder().Value(resultPrefix+method).Continues(true).Build())
		}
		return ret
	})
}

func main() {
	compmain.Main(spec)
}

// TODO Please someone write a formatter...

var spec = "//" + compromise.NewDirectives().SetSourceLocation().Tab(4).JSON() + `
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

@command m :m
@command mm :mm
@command mmm :mmm
@command mma :mm
@command mmma :mmm

@command runahat    :runahat
@command stacktrace :stacktrace
@command stacks     :stacktrace

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

			@call :intent_body_activity

		start-service|start-foreground-service|stop-service		# Start/stop a service.
			@switchloop "^-"
				@call :intent_flags
				@call :take_user_id

			@call :intent_body_service

		broadcast		# Send a broadcast.
			@switchloop "^-"
				@call :intent_flags
				@call :take_user_id

			@call :intent_body_receiver

		dumpheap		# Dump the heap of a process.
			@switchloop "^-"
				-n		# dump native heap instead of managed heap
				-g		# force GC before dumping the heap
				@call :take_user_id

			@switch
				@cand takeProcessName

			@cand takeFile

		instrument 	#  Start an Instrumentation.
			@switchloop "^-"
	      		-r		# print raw results (otherwise decode REPORT_KEY_STREAMRESULT).  Use with \
						#[-e perf true] to generate raw output for performance measurements.
	      		-e 		# <NAME> <VALUE>: set argument <NAME> to <VALUE>.  For test runners a \
	          			# common form is [-e <testrunner_flag> <value>[,<value>...]].
					@any #NAME
					@any #VALUE
	      		-p		# <FILE>: write profiling data to <FILE>
					@cand takeFile
	      		-m		# Write output as protobuf (machine readable)
	      		-w		# wait for instrumentation to finish before returning.  Required for test runners.
				@call :take_user_id
				--no-window-animation  # turn off window animations while running.
	      		--abi 	# <ABI>: Launch the instrumented process with the selected ABI.

			@cand takeDeviceInstrumentation

		trace-ipc	#Trace IPC transactions.
			@switch
				start
				stop
			@switchloop "^-"
				--dump-file
					@cand takeFile

		profile	# Start and stop profiler on a process.
			@switch
				start
				stop
			@switchloop "^-"
				@call :take_user_id
				--sampling # INTERVAL: use sample profiling with INTERVAL microseconds \
          				#between samples
					@any #INTERVAL
      			--streaming # stream the profiling output to the specified file

			@cand takeProcessName
			@cand takeFile

		force-stop # Completely stop the given application package.
			@call :take_user_id
			@cand takeDevicePackage

		kill # Kill all processes associated with the given application.
			@call :take_user_id
			@cand takeDevicePackage

		make-uid-idle 	# If the given application's uid is in the background and waiting to \
      					# become idle (not allowing background services), do that now.
			@call :take_user_id
			@cand takeDevicePackage

		kill-all # Kill all processes that are safe to kill (cached, etc).

		crash 		# Induce a VM crash in the specified package or process
			@call :take_user_id
			@switch
				@cand takeDevicePackage
				@cand takePid

		watch-uids		# Start watching for and reporting uid state changes.
			@switchloop "^-"
      			--oom		# specify a uid for which to report detailed change messages.
			@any #UID // TODO

		get-uid-state 		# Gets the process state of an app given its <UID>.
			@any #UID // TODO

		hang 		# Hang the system.
      		--allow-restart	# allow watchdog to perform normal system restart

  		restart		# Restart the user-space system.

		idle-maintenance # Perform idle maintenance now.

		package-importance 	# Print current importance of <PACKAGE>.
				@cand takeDevicePackage

		switch-user|start-user|unlock-user 	# Switch/start/unlock a user
			@cand takeUserID

		stop-user 	# Stop a user
			@switchloop "^-"
				-w	# wait for stop-user to complete.
				-f	# force stop even if there are related users that cannot be stopped.
			@cand takeUserID

		write	# Write all pending state to storage.

		get-standby-bucket		# Returns the standby bucket of an app.
			@call :take_user_id
			@cand takeDevicePackage

		set-standby-bucket		# Puts an app in the standby bucket.
			@call :take_user_id
			@cand takeDevicePackage
			@switch
				active|working_set|frequent|rare

		send-trim-memory		# Send a memory trim event to a <PROCESS>.  May also supply a raw trim int level.
			@call :take_user_id
			@cand takeProcessName
			@switch
				HIDDEN|RUNNING_MODERATE|BACKGROUND|RUNNING_LOW|MODERATE|RUNNING_CRITICAL|COMPLETE

@label :pm
	@switch
		dump	# dump package
			@cand takeDevicePackage

		clear		# Clear app data
			@call :take_user_id
			@cand takeDevicePackageComponent

		enable|disable|disable-user|disable-until-used|default-state|suspend|unsuspend|set-home-activity # Change package state
			@call :take_user_id
			@cand takeDevicePackageComponent

		dump-profiles # Dumps method/class profile files to /data/misc/profman/TARGET-PACKAGE.txt
			@cand takeDevicePackage

		reconcile-secondary-dex-files	# Reconciles the package secondary dex files with the generated oat files.
			@cand takeDevicePackage

		list				# List information
			@switch
				features		# Prints all features of the system.
				instrumentation # Prints all test packages; optionally only those targeting TARGET-PACKAGE
					@switchloop "^-"
	      				-f		# dump the name of the .apk file containing the test package
					@cand takeDevicePackage

				libraries			# Prints all system libraries.
				permission-groups	# Prints all known permission groups.

	  			packages		#Prints all packages; optionally only those whose name contains
					@switchloop "^-"
						-f		# see their associated file
						-d		# filter to only show disabled packages
						-e		# filter to only show enabled packages
						-s		# filter to only show system packages
						-3		# filter to only show third party packages
						-i		# see the installer for the packages
						-l		# ignored (used for compatibility with older releases)
						-U		# also show the package UID
						-u		# also include uninstalled packages
						--uid   # UID: filter to only show packages with the given UID
							@any # UID #TODO
						@call :take_user_id
					@cand takeDevicePackage

				permissions 	# Prints all known permissions; optionally only those in GROUP.
					@switchloop "^-"
						-g # organize by group
						-f # print all information
						-s # short summary
						-d # only list dangerous permissions
						-u # list only the permissions users will see
					@any # Permission group // TODO
			
		grant|revoke # Grant/revoke permission
			@call :take_user_id
			@cand takeDevicePackage
			@cand takePermission



	// TODO Implement other commands...

@label :dumpsys
	@switch
		activity	# Activity Manager dumpsys
			@call :dumpsys-activity
		package		# Package Manager dumpsys
			@call :dumpsys-package
		@cand takeService


@label :dumpsys-activity
	@switch
		activities
		broadcasts
			history
		intents
		lastanr
		permissions
		processes
		providers
		recents
		services
		starter

@label :dumpsys-package
	@switch
		@cand takeDevicePackage


@label :cmd
	@cand takeService

@label :settings
	@switch
		get			# Retrieve the current value of KEY
			@call :take_user_id
			@call :settings_namespace
			@cand takeSettingKey

		put			# Change the contents of KEY to VALUE
			@call :take_user_id
			@call :settings_namespace
			@cand takeSettingKey
			@any	# <value> value to set
			@any	# <tag>
			@switch
				default # {default} to set as the default, case-insensitive only for global/secure namespace

		delete		# Delete the entry for KEY
			@call :settings_namespace
			@cand takeSettingKey

		reset		# Reset the global/secure table for a package with mode
			@call :take_user_id
			@call :settings_namespace
			@switch
				@cand takeDevicePackage
				untrusted_defaults
				untrusted_clear
				trusted_defaults

		list	# Print all defined keys
			@call :settings_namespace

// --user [ N | current | all ] NOT not all commands will understand "current" and "all".
@label :take_user_id
	@switch "^-"
		--user # Specify user-id.
			@switch
				@cand takeUserID
					@go_call setUserId
				current|all
					@go_call setUserId

// settings global put " [ global | system | secure ] "
@label :settings_namespace
	@switch
		global|system|secure
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
	@cand takeDevicePackageComponent

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

@label :intent_body_activity
	@cand takeDeviceActivity

@label :intent_body_service
	@cand takeDeviceService

@label :intent_body_receiver
	@cand takeDeviceReceiver


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
		@cand takeJavaFileMethod	// path/to/filename.java#method1,method2,...

		@cand takeBuildModule "(Test|^Bug)"

@label :m
	@call :makeFlags
	@switchloop
		droid
		installclean
		showcommands
		snod
		vnod
		checkbuild
		cts
		update-api

        @cand takeBuildModule
        @cand takeFile

@label :mm
	@call :makeFlags

@label :mmm
	@call :makeFlags
	@loop
		@cand takeDir


@label :makeflags
	@switchloop "^-"
		-j|--jobs					# Specifies the number of jobs (commands) to run simultaneously.
			@switch "^[0-9]"
				@cand takeInteger
		-i|--ignore-errors 			# Ignore all errors in commands executed to remake files.

@label :runahat
	@switch
		@cand takeProcessName
		@cand takePid

@label :stacktrace
	@switch
		@cand takeProcessName
		@cand takePid
`
