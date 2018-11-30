package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	"github.com/stefanhans/cloud-function-play/firestore/chat/chat-group"
)

// Start service to receive messages
func startChatListener() error {

	// Create chatListener
	chatListener, err := net.Listen(gcpMemberList.Self.Protocol, gcpMemberList.Self.Ip+":"+gcpMemberList.Self.Port)

	if err != nil {
		log.Fatalf("could not listen to %q: %v\n", gcpMemberList.Self.Ip+":"+gcpMemberList.Self.Port, err)
	}

	// Set the port number of the listener as member port
	_, port, err := net.SplitHostPort(chatListener.Addr().String())
	if err != nil {
		fmt.Printf("cannot split host from (new) port %q: %v", chatListener.Addr().String(), err)
	}
	gcpMemberList.Self.Port = port

	defer chatListener.Close()

	log.Printf("Started displaying service listening on %q\n", gcpMemberList.Self.Ip+":"+gcpMemberList.Self.Port)

	for {
		// Wait for a connection.
		conn, err := chatListener.Accept()
		if err != nil {
			continue //log.Fatal(err)
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handleMessage(conn)
	}

	return nil
}

var messageActionMap = map[chatgroup.Message_MessageType]func(*chatgroup.Message) error{
	chatgroup.Message_SUBSCRIBE_REPLY:   handleSubscribeReply,
	chatgroup.Message_UNSUBSCRIBE_REPLY: handleUnsubscribeReply,
	chatgroup.Message_PUBLISH_REPLY:     handlePublishReply,

	chatgroup.Message_TEST_PUBLISH_REQUEST: handleTestPublishRequest,
	chatgroup.Message_TEST_CMD_REQUEST:     handleTestCmdRequest,
}

// Read all incoming data, take the leading byte as message type,
// and use the appropriate handler for the rest
func handleMessage(conn net.Conn) {

	defer conn.Close()

	// Read all data from the connection
	var buf [512]byte
	var data []byte
	addr := conn.RemoteAddr()

	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			break
		}
		data = append(data, buf[0:n]...)
	}

	//log.Printf("received %v bytes\n", len(data))

	// Unmarshall message
	var msg chatgroup.Message
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		fmt.Printf("could not unmarshall message: %v", err)
	}

	jsonMsg, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal message: %v\n", err)
	}
	cLog.Printf("FROM %s: \n%s", addr.String(), string(jsonMsg))

	// Fetch the handler from a map by the message type and call it accordingly
	if replyAction, ok := messageActionMap[msg.MsgType]; ok {
		err := replyAction(&msg)
		if err != nil {
			log.Printf("could not handle %v from %v: %v", msg.MsgType, addr, err)
		}
	} else {
		log.Printf("handleMessage: unknown message type %v\n", msg.MsgType)
	}
}

// handleSubscribeReply inserts new member into the chat's member list idempotentially
func handleSubscribeReply(msg *chatgroup.Message) error {
	log.Printf("handleSubscribeReply(%v)\n", msg)

	// Append text message in "messages" view
	displayText(fmt.Sprintf("<%s (%s:%s) has joined>", msg.Sender.Name, msg.Sender.Ip, msg.Sender.Port))

	// Skip subscriber if already exists
	for _, m := range chat.memberlist {
		if m.Name == msg.Sender.Name {
			return nil
		}
	}
	// Add subscriber
	chat.memberlist = append(chat.memberlist, msg.Sender)

	return nil
}

// handleUnsubscribeReply removes the member from the chat's member list
func handleUnsubscribeReply(msg *chatgroup.Message) error {

	// Append text message in "messages" view
	displayText(fmt.Sprintf("<%s has left>", msg.Sender.Name))

	// Remove subscriber
	for i, s := range chat.memberlist {
		if s.Name == msg.Sender.Name {
			chat.memberlist = append(chat.memberlist[:i], chat.memberlist[i+1:]...)
			return nil
		}
	}

	return nil
}

// handlePublishReply displays the text in the chat
func handlePublishReply(msg *chatgroup.Message) error {
	log.Printf("handlePublishReply(%v)\n", msg)

	// Append text message in "messages" view
	displayText(fmt.Sprintf("%s: %s", msg.Sender.Name, msg.Text))

	return nil
}
