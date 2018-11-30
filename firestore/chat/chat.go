package main

import (
	"fmt"

	"github.com/stefanhans/cloud-function-play/firestore/chat/chat-group"
	gcp_memberlist "github.com/stefanhans/cloud-function-play/memberlist"
)

// Chat is the core struct for the chat
type Chat struct {
	self       *chatgroup.Member
	memberlist []*chatgroup.Member
	message    *chatgroup.Message
}

// CreateChat returns a new chat instance
func CreateChat(name, ip string) *Chat {

	self := &chatgroup.Member{
		Name: name,
		Ip:   ip,
	}

	var memberlist []*chatgroup.Member

	return &Chat{
		self:       self,
		memberlist: memberlist,
		message: &chatgroup.Message{
			MsgType: chatgroup.Message_SUBSCRIBE_REQUEST,
			Sender:  self,
		},
	}
}

// Implemets the Stringer interface
func (chat *Chat) String() string {
	out := fmt.Sprintf("self: %v\n", chat.self)
	out += fmt.Sprintf("memberlist: %v\n", chat.memberlist)
	out += fmt.Sprintf("message: %v", chat.message)

	return out
}

// Initialize synchronizes the loaded member list with this chat and informs the others
func (chat *Chat) Initialize(gcpList map[string]*gcp_memberlist.IpAddress) error {

	if len(gcpList) == 0 {
		return fmt.Errorf("empty memberlist\n")
	}

	// Convert Ip adresses from memberlist
	for _, v := range gcpList {
		chat.memberlist = append(chat.memberlist, &chatgroup.Member{
			Name:     v.Name,
			Ip:       v.Ip,
			Port:     v.Port,
			Protocol: v.Protocol,
		})
	}

	// Set the own Ip address from memberlist
	chat.self = &chatgroup.Member{
		Name:     gcpMemberList.Self.Name,
		Ip:       gcpMemberList.Self.Ip,
		Port:     gcpMemberList.Self.Port,
		Protocol: gcpMemberList.Self.Protocol,
	}
	chat.message.Sender = chat.self

	// Publish list of subscribers
	err = chat.publishSubscriberList()
	if err != nil {
		return fmt.Errorf("failed to publish subscriber list: %v\n", err)
	}

	return nil
}

// Publish the member list to all members - the sender included
func (chat *Chat) publishSubscriberList() error {

	chat.message.MsgType = chatgroup.Message_SUBSCRIBE_REPLY

	// Send message to all chat group members
	for _, recipient := range chat.memberlist {
		err := sendMessage(chat.message, recipient.Ip+":"+recipient.Port)
		if err != nil {
			return fmt.Errorf("failed send subscription to %v:%v: %v", recipient.Ip, recipient.Port, err)
		}
	}
	return nil
}
