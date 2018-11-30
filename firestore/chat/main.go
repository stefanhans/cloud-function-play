package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jroimartin/gocui"
)

// todo: establish a datastructure for a chat instance

var (
	chat *Chat
	err  error
)

func main() {

	// Check command args and set own chatgroup.Member
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "missing or wrong parameter: <name> <chat ip>")
		os.Exit(1)
	}

	// Start logging into one file each session
	logfile, err = startLogging(flag.Arg(0))
	if err != nil {
		log.Fatalf("error starting logging: %v", err)
	}
	defer logfile.Close()

	// Create memberlist for GCP Cloud Functions with Firestore
	gcpMemberList, err = CreateMemberlist(flag.Arg(0), flag.Arg(1))
	if err != nil {
		log.Fatalf("error creating memberlist: %v", err)
	}

	// Start logging into one common file for all sessions of today
	cLog, commonLogfile, err = startCommonLogging(flag.Arg(0), gcpMemberList.Uuid)
	if err != nil {
		log.Fatalf("error starting logging: %v", err)
	}
	defer commonLogfile.Close()

	// Set the current ChatListener.Port and wait for its update after the start
	currentChatListenerPort := gcpMemberList.Self.Port

	go func() {

		err = startChatListener()
		if err != nil {
			log.Fatalf("Failed to start chat listener on \"%s:%s\": %v", gcpMemberList.Self.Ip, gcpMemberList.Self.Port, err)
		}
	}()

	for gcpMemberList.Self.Port == currentChatListenerPort {
		time.Sleep(time.Millisecond * 10)
	}

	// Subscribe at and get the memberlist provided by GCP service
	gcpList, err := gcpMemberList.Subscribe()
	if err != nil {
		log.Fatalf("error loading memberlist from GCP service: %v", err)
	}

	// Create the TUI
	clientGui, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatalf("could not create tui: %v\n", err)
	}
	defer clientGui.Close()

	// Create and initialize the chat instance
	chat = CreateChat(flag.Arg(0), flag.Arg(1))
	err = chat.Initialize(gcpList)
	if err != nil {
		log.Fatalf("error initializing chat: %v", err)
	}

	// Initialize chat command usage
	commandUsageInit()

	// Start text-based UI
	err = runTUI()
	if err != nil {
		log.Fatalf("runTUI: %v", err)
	}
}
