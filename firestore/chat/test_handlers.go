package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/stefanhans/cloud-function-play/firestore/chat/chat-group"
)

// handleTestPublishRequest sends the test message firstly to all members,
// and if receiving this kind of message itself, it sends them back to the tester
func handleTestPublishRequest(msg *chatgroup.Message) error {

	// Gets the current senders name and the IP address of the tester
	name := msg.Sender.Name
	testerIpAddress := strings.Split(msg.Text, "|")[0]

	// Coming firstly from tester
	if msg.Sender.Ip+":"+msg.Sender.Port == testerIpAddress {

		// Update sender and append to test message text
		msg.Sender = &chatgroup.Member{
			Name:     gcpMemberList.Self.Name,
			Ip:       gcpMemberList.Self.Ip,
			Port:     gcpMemberList.Self.Port,
			Protocol: gcpMemberList.Self.Protocol,
		}
		msg.Text += "|" + gcpMemberList.Self.Name + ":" + gcpMemberList.Self.Ip + ":" + gcpMemberList.Self.Port

		// Send to all members - including itself
		sendMessageToAll(msg)

		// Shows information about being tested in the chat
		displayText(fmt.Sprintf("<TEST_PUBLISH %q>", name))

	} else {

		// Update sender and append to test message text
		msg.Sender = &chatgroup.Member{
			Name:     gcpMemberList.Self.Name,
			Ip:       gcpMemberList.Self.Ip,
			Port:     gcpMemberList.Self.Port,
			Protocol: gcpMemberList.Self.Protocol,
		}
		msg.Text += "|" + gcpMemberList.Self.Name + ":" + gcpMemberList.Self.Ip + ":" + gcpMemberList.Self.Port

		// Change message type and send back to tester
		msg.MsgType = chatgroup.Message_TEST_PUBLISH_REPLY
		sendMessage(msg, testerIpAddress)

		// Shows information about being tested in the chat
		displayText(fmt.Sprintf("<TEST_REPLY %q>", name))
	}

	return nil
}

// handleTestCmdRequest sends the test message with a command,
// and it sends the command's result back to the tester
func handleTestCmdRequest(msg *chatgroup.Message) error {

	// Shows information about being tested in the chat
	displayText(fmt.Sprintf("<TEST_CMD %q>", msg.Text))

	// Save the sender's address for the reply and updates the message's sender
	testerIpAddress := msg.Sender.Ip + ":" + msg.Sender.Port
	msg.Sender = &chatgroup.Member{
		Name:     gcpMemberList.Self.Name,
		Ip:       gcpMemberList.Self.Ip,
		Port:     gcpMemberList.Self.Port,
		Protocol: gcpMemberList.Self.Protocol,
	}

	// Todo: Refactor the function to close it for changes while keeping it open to extensions

	// Switch due to the command in the message text and act accordingly
	switch msg.Text {

	case "list":

		// Append the chat's member list as JSON at the message text delimited by a pipe symbol
		jsonChatMemberlist, err := json.MarshalIndent(chat.memberlist, "", "  ")
		if err != nil {
			log.Printf("failed to marshal chat's member list: %v\n", err)
		}
		msg.Text += "|" + string(jsonChatMemberlist)

	default:
		log.Printf("unknown command to test: %v\n", msg.Text)
	}

	// Change message type and send back to tester
	msg.MsgType = chatgroup.Message_TEST_CMD_REPLY
	sendMessage(msg, testerIpAddress)

	return nil
}

//func writeTest(g *gocui.Gui, v *gocui.View, txt string) error {
//	cLog.Printf("writeTest(...): \n")
//
//	v.Clear()
//	v.Write([]byte(txt))
//	send(g, v)
//
//	return nil
//}
