package main

import (
	"fmt"
	"log"
	"net"

	"github.com/stefanhans/cloud-function-play/firestore/chat/chat-group"

	"github.com/golang/protobuf/proto"
)

// Unsubscribe calls the subscription service and informs all other chat members
func Unsubscribe() error {

	err := gcpMemberList.Unsubscribe()
	if err != nil {
		log.Fatalf("error unsubscribing from memberlist of GCP service: %v", err)
	}

	// Remove subscriber
	for i, s := range chat.memberlist {
		if s.Name == chat.self.Name {
			chat.memberlist = append(chat.memberlist[:i], chat.memberlist[i+1:]...)
			break
		}
	}

	message := &chatgroup.Message{
		MsgType: chatgroup.Message_UNSUBSCRIBE_REPLY,
		Sender: &chatgroup.Member{
			Name: chat.self.Name,
		},
	}
	err = sendMessageToAll(message)
	if err != nil {
		return err
	}

	return nil
}

func Publish(text string) error {
	cLog.Printf("Publish(%q)\n", text)

	chat.message.MsgType = chatgroup.Message_PUBLISH_REPLY
	chat.message.Text = text

	return sendMessageToAll(chat.message)
}

// sendMessageToAll sends the message to all members hold in the chat's memberlist
func sendMessageToAll(message *chatgroup.Message) error {

	// Forward message to other chat group members
	for _, recipient := range chat.memberlist {
		// Send message to recipient
		err := sendMessage(message, recipient.Ip+":"+recipient.Port)
		if err != nil {
			return fmt.Errorf("Failed send reply: %v", err)
		}
	}
	return nil
}

// Send reply to the sender of the message
func sendMessage(msg *chatgroup.Message, recipient string) error {
	cLog.Printf("sendMessage(%v, %s)\n", msg, recipient)

	// Connect to the recipient
	conn, err := net.Dial("tcp", recipient)
	if err != nil {
		return fmt.Errorf("could not connect to recipient %q: %v", recipient, err)
	}

	// Marshal into binary format
	byteArray, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("could not encode message: %v", err)
	}

	// Write the bytes to the connection
	_, err = conn.Write(byteArray)
	if err != nil {
		return fmt.Errorf("could not write message to the connection: %v", err)
	}

	// Close connection
	return conn.Close()
}
