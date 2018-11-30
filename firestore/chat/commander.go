package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	gcp_memberlist "bitbucket.org/stefanhans/go-thesis/Libraries/memberlist"
	"github.com/stefanhans/cloud-function-play/firestore/chat/chat-group"
)

var (
	cmdUsage map[string]string
	keys     []string
)

func commandUsageInit() {
	cmdUsage = make(map[string]string)

	cmdUsage["all"] = "\\all"

	cmdUsage["chat"] = "\\chat"
	cmdUsage["self"] = "\\self"
	cmdUsage["list"] = "\\list"
	cmdUsage["message"] = "\\message"
	cmdUsage["logfile"] = "\\logfile"

	cmdUsage["gcp"] = "\\gcp"
	cmdUsage["gcpconfig"] = "\\gcpconfig"
	cmdUsage["gcplist"] = "\\gcplist"

	cmdUsage["gcpreset"] = "\\gcpreset"
	cmdUsage["gcpsubscribe"] = "\\gcpsubscribe"
	cmdUsage["gcpunsubscribe"] = "\\gcpunsubscribe"

	cmdUsage["types"] = "\\types"

	cmdUsage["quit"] = "\\quit"

	// To store the keys in sorted order
	for key := range cmdUsage {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	log.Printf("commandUsageInit: keys: %v\n", keys)
}

// Execute a command specified by the argument string
func executeCommand(commandline string) {

	// Trim prefix and split string by white spaces
	commandFields := strings.Fields(strings.Trim(commandline, "\\"))

	// Check for empty string without prefix
	if len(commandFields) > 0 {
		log.Printf("Command: %q\n", commandFields[0])
		log.Printf("Arguments (%v): %v\n", len(commandFields[1:]), commandFields[1:])

		// Switch according to the first word and call appropriate function with the rest as arguments
		switch commandFields[0] {

		case "all":
			// enhancement: order deterministic
			log.Printf("CMD_ALL\n")
			self(commandFields[1:])
			list(commandFields[1:])
			message(commandFields[1:])
			showLogfile(commandFields[1:])
			gcpconfig(commandFields[1:])
			gcplist(commandFields[1:])

		// CHAT
		case "chat":
			// enhancement: order deterministic
			log.Printf("CMD_CHAT\n")
			self(commandFields[1:])
			list(commandFields[1:])
			message(commandFields[1:])
			showLogfile(commandFields[1:])

		case "self":
			log.Printf("CMD_SELF\n")
			self(commandFields[1:])

		case "list":
			log.Printf("CMD_LIST\n")
			list(commandFields[1:])

		case "message":
			log.Printf("CMD_MESSAGE\n")
			list(commandFields[1:])

		case "logfile":
			log.Printf("CMD_LOGFILE\n")
			showLogfile(commandFields[1:])

		// GCP
		case "gcp":
			log.Printf("CMD_GCP\n")
			gcpconfig(commandFields[1:])
			gcplist(commandFields[1:])

		case "gcpconfig":
			log.Printf("CMD_GCP_CONFIG\n")
			gcpconfig(commandFields[1:])

		case "gcplist":
			log.Printf("CMD_GCP_LIST\n")
			gcplist(commandFields[1:])

		// DEBUG
		case "gcpreset":
			log.Printf("CMD_GCP_RESET\n")
			gcpreset(commandFields[1:])

		case "gcpsubscribe":
			log.Printf("CMD_GCP_SUBSCRIBE\n")
			gcpsubscribe(commandFields[1:])

		case "gcpunsubscribe":
			log.Printf("CMD_GCP_UNSUBSCRIBE\n")
			gcpunsubscribe(commandFields[1:])

		case "types":
			log.Printf("CMD_TYPES\n")
			types(commandFields[1:])

		case "quit":
			log.Printf("CMD_GCP_UNSUBSCRIBE\n")
			quitChat(commandFields[1:])

		default:
			usage()
		}

	} else {
		usage()
	}
}

// Display the usage of all available commands
func usage() {
	// enhance: order not deterministic bug
	for _, key := range keys {
		displayText(fmt.Sprintf("<CMD USAGE>: %s", cmdUsage[key]))
	}
}

func message(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	jsonChatMessage, err := json.MarshalIndent(chat.message, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal chat.message: %v\n", err)
	}

	displayText(strings.Trim(fmt.Sprintf("<CMD_MESSAGE>: \n%v\n%s", string(jsonChatMessage), last), "\n"))
}

func list(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	jsonChatMemberlist, err := json.MarshalIndent(chat.memberlist, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal chat.memberlist: %v\n", err)
	}

	displayText(strings.Trim(fmt.Sprintf("<CMD_LIST>: \n%v\n%s", string(jsonChatMemberlist), last), "\n"))
}

func showLogfile(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	jsonLogfilename, err := json.MarshalIndent(logfilename, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal logfilename: %v\n", err)
	}

	displayText(fmt.Sprintf("<CMD_LOGFILE>: \n%v\n%s", string(jsonLogfilename), last))
}

func self(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	jsonSelf, err := json.MarshalIndent(gcpMemberList.Self, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal gcpMemberList.Self: %v\n", err)
	}

	displayText(strings.Trim(fmt.Sprintf("<CMD_SELF>: \n%v\n%s", string(jsonSelf), last), "\n"))
}

func gcpconfig(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	jsonGcpMemberList, err := json.MarshalIndent(gcpMemberList, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal gcpMemberList: %v\n", err)
	}

	displayText(strings.Trim(fmt.Sprintf("<CMD_GPC_CONFIG>: \n%v\n%s", string(jsonGcpMemberList), last), "\n"))
}

func gcplist(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	gcpList, err := gcpMemberList.List()
	if err != nil {
		displayText(fmt.Sprintf("<CMD_GCP_LIST>: List() call failed: %v\n", err))
		return
	}

	if len(gcpList) == 0 {
		displayText(fmt.Sprintf("<CMD_GCP_LIST>: empty\n"))
		return
	}

	jsonGcpList, err := json.MarshalIndent(gcpList, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal gcp memberlist list: %v\n", err)
	}

	displayText(strings.Trim(fmt.Sprintf("<CMD_GCP_LIST>: \n%v\n%s", string(jsonGcpList), last), "\n"))
}

func gcpreset(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	err := gcp_memberlist.Reset(gcpMemberList.ServiceUrl)
	if err != nil {
		displayText(fmt.Sprintf("<CMD_GCP_RESET>: Reset() call failed: %v\n", err))
		return
	}

	displayText(fmt.Sprintf("<CMD_GCP_RESET>: done\n%s", last))
}

func gcpsubscribe(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	gcpList, err := gcpMemberList.Subscribe()
	if err != nil {
		displayText(fmt.Sprintf("<CMD_GCP_SUBSCRIBE>: Subscribe() call failed: %v\n", err))
		return
	}

	chat.memberlist = append(chat.memberlist[:0], chat.memberlist[:0]...)

	for _, v := range gcpList {
		chat.memberlist = append(chat.memberlist, &chatgroup.Member{
			Name:     v.Name,
			Ip:       v.Ip,
			Port:     v.Port,
			Protocol: v.Protocol,
		})
	}

	err = chat.publishSubscriberList()
	if err != nil {
		displayText(fmt.Sprintf("<CMD_GCP_SUBSCRIBE>: publishSubscriberList() call failed: %v\n", err))
		return
	}

	displayText(fmt.Sprintf("<CMD_GCP_SUBSCRIBE>: done\n%s", last))
}

func gcpunsubscribe(arguments []string) {

	// Append arguments for distributed testing
	last := strings.Join(arguments, " ")

	err := gcpMemberList.Unsubscribe()
	if err != nil {
		displayText(fmt.Sprintf("<CMD_GCP_UNSUBSCRIBE>: Unsubscribe() call failed: %v\n", err))
	}

	displayText(fmt.Sprintf("<CMD_GCP_UNSUBSCRIBE>: done\n%s", last))
}

func types(arguments []string) {

	// Get rid of warnings
	_ = arguments

	txt := "<CMD_TYPES>: SUBSCRIBE_REQUEST =    0; \n"
	txt += "<CMD_TYPES>: SUBSCRIBE_REPLY =      1; \n"
	txt += "<CMD_TYPES>: UNSUBSCRIBE_REQUEST =  2; \n"
	txt += "<CMD_TYPES>: UNSUBSCRIBE_REPLY =    3; \n"
	txt += "<CMD_TYPES>: PUBLISH_REQUEST =      4; \n"
	txt += "<CMD_TYPES>: PUBLISH_REPLY =        5; \n"
	txt += "<CMD_TYPES>: TEST_PUBLISH_REQUEST = 6; \n"
	txt += "<CMD_TYPES>: TEST_PUBLISH_REPLY =   7; \n"
	txt += "<CMD_TYPES>: TEST_CMD_REQUEST =     8; \n"
	txt += "<CMD_TYPES>: TEST_CMD_REPLY =       9; \n"

	displayText(txt)
}

func quitChat(arguments []string) {

	// Get rid of warnings
	_ = arguments

	inputView, _ := clientGui.View("input")

	quit(clientGui, inputView)

	os.Exit(0)
}
