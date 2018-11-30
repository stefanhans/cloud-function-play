package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
	"github.com/stefanhans/cloud-function-play/firestore/chat/chat-group"
)

var (
	// IpAddress for the listener of the test
	testingIpAddress = "127.0.0.1:22365"

	// Map with message types as keys and appropriate reply handlers as values
	replyTestMap = map[chatgroup.Message_MessageType]func(*chatgroup.Message) error{
		chatgroup.Message_TEST_PUBLISH_REPLY: handleTestPublishReply,
		chatgroup.Message_TEST_CMD_REPLY:     handleTestCmdReply,
	}

	// Channels to handle the returning messages of the tests
	testPublishReplyChannel = make(chan *chatgroup.Message)
	listReplyChannel        = make(chan *chatgroup.Message)
)

// startTestListener provides a listener for returning test messages
func startTestListener() error {

	testingListener, err := net.Listen("tcp", testingIpAddress)
	if err != nil {
		log.Fatalf("could not listen to %q: %v\n", testingIpAddress, err)
	}
	defer testingListener.Close()

	for {
		// Wait for a connection.
		conn, err := testingListener.Accept()
		if err != nil {
			continue
		}

		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go handleTesterReply(conn)
	}

	return nil
}

// Read all incoming data, take the leading byte as message type,
// and use the appropriate handler for the rest
func handleTesterReply(conn net.Conn) {

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

	// Unmarshall message
	var msg chatgroup.Message
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		fmt.Printf("could not unmarshall message: %v", err)
	}

	// Fetch the handler from a map by the message type and call it accordingly
	if replyTest, ok := replyTestMap[msg.MsgType]; ok {
		//log.Printf("%v\n", msg)
		err := replyTest(&msg)
		if err != nil {
			fmt.Printf("could not handle %v from %v: %v", msg.MsgType, addr, err)
		}
	} else {
		fmt.Printf("testing: unknown message type %v\n", msg.MsgType)
	}
}

// TestMain is the initial function for the tests
func TestMain(m *testing.M) {
	go func() {
		err = startTestListener()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed start test listener: %v", err)
			os.Exit(1)
		}
	}()
	fmt.Printf("Test service listening on %q\n", testingIpAddress)
	os.Exit(m.Run())
}

// handleTestPublishReply is the registered handler for TEST_PUBLISH_REPLY messages
func handleTestPublishReply(msg *chatgroup.Message) error {

	// Sends the message into the channel
	testPublishReplyChannel <- msg

	return nil
}

// TestPublishMessage sends test messages to all members and expect messages from all possible permutations
func TestPublishMessage(t *testing.T) {

	testname := "TestPublishMessage"

	// Create memberlist for GCP Cloud Functions with Firestore
	gcpMemberList, err = CreateMemberlist("tester", "127.0.0.1")
	if err != nil {
		t.Fatalf("error creating memberlist: %v", err)
	}

	// Get current list of members
	gcpList, err := gcpMemberList.List()
	if err != nil {
		t.Fatalf("error creating memberlist: %v", err)
	}

	if len(gcpList) == 0 {
		t.Fatalf("empty memberlist: %v", err)
	}

	// Create message to start tests
	id := uuid.NewV4().String()
	msg := &chatgroup.Message{
		MsgType: chatgroup.Message_TEST_PUBLISH_REQUEST,
		Sender: &chatgroup.Member{
			Name:     testname,
			Ip:       "127.0.0.1",
			Port:     "22365",
			Protocol: "tcp",
		},
		Text: testingIpAddress + "|" + testname + ":" + id,
	}

	// Send test message to all members of the list
	testInitCnt := 0
	for _, member := range gcpList {

		recipient := member.Ip + ":" + member.Port

		// Get the connection to the recipient
		conn, err := net.Dial("tcp", recipient)
		if err != nil {
			t.Errorf("could not connect to recipient %q: %v", recipient, err)
		}

		// Marshal into binary format
		byteArray, err := proto.Marshal(msg)
		if err != nil {
			t.Errorf("could not encode message: %v", err)
		}

		// Write the bytes to the connection
		_, err = conn.Write(byteArray)
		if err != nil {
			t.Errorf("could not write message to the connection: %v", err)
		}

		// Close connection
		conn.Close()

		// Count number of test initialized
		testInitCnt++
	}

	// Number of permutations and expected messages, respectively
	testInitCnt *= testInitCnt

	// Waiting until all messages arrived
	testResultCnt := 1
	senderCnt, receiverCnt := 0, 0
	for {
		select {
		case testResult := <-testPublishReplyChannel:

			// Prepare one incoming test message
			resultParts := strings.Split(testResult.Text, "|")

			// Check the two first parts, i.e. the test header
			if resultParts[0] != testingIpAddress {
				t.Errorf("not the expected testingIpAddress (1|...): %s\n", testResult.Text)
			}
			if resultParts[1] != testname+":"+id {
				t.Errorf("not the expected testname:uuid (...|2|...): %s\n", testResult.Text)
			}

			// Count valid sender and receiver
			for _, member := range gcpList {
				if fmt.Sprintf("%v:%v:%v", member.Name, member.Ip, member.Port) == resultParts[2] {
					senderCnt++
				}
				if fmt.Sprintf("%v:%v:%v", member.Name, member.Ip, member.Port) == resultParts[3] {
					receiverCnt++
				}
			}
		}

		// Todo: implement done channel for timeout
		if testResultCnt == testInitCnt {
			break
		}
		testResultCnt++
	}

	// Check the message and hop counts
	if testResultCnt != testInitCnt {
		t.Errorf("only %d of expected %d messages exist\n", testResultCnt, testInitCnt)
	}
	if senderCnt != testInitCnt {
		t.Errorf("only %d of expected %d messages sent\n", senderCnt, testInitCnt)
	}
	if receiverCnt != testInitCnt {
		t.Errorf("only %d of expected %d messages received\n", receiverCnt, testInitCnt)
	}
}

// handleTestCmdReply is the registered handler for TEST_CMD_REPLY messages
// and it multiplexes the test result into the appropriate channel
func handleTestCmdReply(msg *chatgroup.Message) error {

	// Prepare the message text, i.e. seperate the command from its returned data
	msgTexts := strings.SplitN(msg.Text, "|", 2)

	// Todo: Refactor the function to close it for changes while keeping it open to extensions

	// Switch due to the command in the message text and forward to the appropriate channel
	switch msgTexts[0] {

	case "list":

		// Sends the returned list into the channel
		msg.Text = msgTexts[1]
		listReplyChannel <- msg

	default:
		fmt.Printf("unknown command for test: %v\n", msgTexts[0])
	}

	return nil
}

func TestCmdList(t *testing.T) {

	testname := "TestCmdList"

	// Create memberlist for GCP Cloud Functions with Firestore
	gcpMemberList, err = CreateMemberlist("tester", "127.0.0.1")
	if err != nil {
		t.Fatalf("error creating memberlist: %v", err)
	}

	// Get current list of members
	gcpList, err := gcpMemberList.List()
	if err != nil {
		t.Fatalf("error creating memberlist: %v", err)
	}

	if len(gcpList) == 0 {
		t.Fatalf("empty memberlist: %v", err)
	}

	// Create message to start tests
	id := uuid.NewV4().String()
	msg := &chatgroup.Message{
		MsgType: chatgroup.Message_TEST_CMD_REQUEST,
		Sender: &chatgroup.Member{
			Name:     testname + ":" + id,
			Ip:       "127.0.0.1",
			Port:     "22365",
			Protocol: "tcp",
		},
		Text: "list",
	}

	// Send test message to all members of the list
	testInitCnt := 0
	for _, member := range gcpList {

		recipient := member.Ip + ":" + member.Port

		// Get the connection to the recipient
		conn, err := net.Dial("tcp", recipient)
		if err != nil {
			t.Errorf("could not connect to recipient %q: %v", recipient, err)
		}

		// Marshal into binary format
		byteArray, err := proto.Marshal(msg)
		if err != nil {
			t.Errorf("could not encode message: %v", err)
		}

		// Write the bytes to the connection
		_, err = conn.Write(byteArray)
		if err != nil {
			t.Errorf("could not write message to the connection: %v", err)
		}

		// Close connection
		conn.Close()

		// Count number of test initialized
		testInitCnt++
	}

	// Convert Ip adresses from member list from the subscription service
	var testMemberlist []*chatgroup.Member
	for _, member := range gcpList {
		testMemberlist = append(testMemberlist, &chatgroup.Member{
			Name:     member.Name,
			Ip:       member.Ip,
			Port:     member.Port,
			Protocol: member.Protocol,
		})
	}

	// Waiting until all messages arrived
	testResultCnt := 1
	for {
		select {
		case testResult := <-listReplyChannel:

			// Unmarshall message's member list
			var receivedMemberlist []*chatgroup.Member
			err = json.Unmarshal([]byte(testResult.Text), &receivedMemberlist)
			if err != nil {
				t.Errorf("could not unmarshall received member list: %v", err)
			}

			// Test the length of the received member list against the expected
			if len(receivedMemberlist) != len(testMemberlist) {
				t.Errorf("only %d of expected %d members are listed at %q\n", len(receivedMemberlist),
					len(testMemberlist), testResult.Sender.Name)
			}

			// Count the matches and test the number against the expected
			matchedMemberCnt := 0
			for _, receivedMember := range receivedMemberlist {
				for _, listedMember := range testMemberlist {
					if fmt.Sprintf("%v", receivedMember) == fmt.Sprintf("%v", listedMember) {
						matchedMemberCnt++
					}
				}
			}
			if len(testMemberlist) != matchedMemberCnt {
				t.Errorf("only %d of expected %d members are listed at %q\n", matchedMemberCnt,
					len(testMemberlist), testResult.Sender.Name)
			}
		}

		// Todo: implement done channel for timeout
		if testResultCnt == testInitCnt {
			break
		}
		testResultCnt++
	}

	// Check the message counts
	if testResultCnt != testInitCnt {
		t.Errorf("only %d of expected %d messages exist\n", testResultCnt, testInitCnt)
	}
}

//func TestQuitChats(t *testing.T) {
//
//	testname := "TestQuitChats"
//
//	// Create memberlist for GCP Cloud Functions with Firestore
//	gcpMemberList, err = CreateMemberlist("tester", "127.0.0.1")
//	if err != nil {
//		t.Fatalf("error creating memberlist: %v", err)
//	}
//
//	gcpList, err := gcpMemberList.List()
//	if err != nil {
//		t.Fatalf("error call memberlist.List(): %v", err)
//	}
//
//	var id string
//	var ipAdress *gcp_memberlist.IpAddress
//
//	var ids []string
//
//	i := 0
//	for k, v := range gcpList {
//
//		id = k
//		ipAdress = v
//
//		ids = append(ids, id)
//
//		message := &chatgroup.Message{
//			MsgType: chatgroup.Message_TEST_CMD_REQUEST,
//			Sender: &chatgroup.Member{
//				Name: testname,
//			},
//			Text: "\\quit",
//		}
//
//		err = sendMessage(message, ipAdress.Ip+":"+ipAdress.Port)
//		if err != nil {
//			t.Errorf("Failed send reply: %v", err)
//		}
//		i++
//	}
//}

//func TestGcpList(t *testing.T) {
//
//	testname := "TestGcpList"
//
//	// Create memberlist for GCP Cloud Functions with Firestore
//	gcpMemberList, err = CreateMemberlist("tester", "127.0.0.1")
//	if err != nil {
//		t.Fatalf("error creating memberlist: %v", err)
//	}
//
//	gcpList, err := gcpMemberList.List()
//	if err != nil {
//		t.Fatalf("error call memberlist.List(): %v", err)
//	}
//
//	var id string
//	var ipAdress *gcp_memberlist.IpAddress
//
//	var ids []string
//
//	i := 0
//	for k, v := range gcpList {
//
//		id = k
//		ipAdress = v
//
//		ids = append(ids, id)
//
//		message := &chatgroup.Message{
//			MsgType: chatgroup.Message_TEST_CMD_REQUEST,
//			Sender: &chatgroup.Member{
//				Name: testname,
//			},
//			Text: "\\gcplist",
//		}
//
//		err = sendMessage(message, ipAdress.Ip+":"+ipAdress.Port)
//		if err != nil {
//			t.Errorf("Failed send reply: %v", err)
//		}
//		i++
//	}
//}

//func TestWork(t *testing.T) {
//
//	testname := "TestWork"
//
//	// Create memberlist for GCP Cloud Functions with Firestore
//	gcpMemberList, err = CreateMemberlist("tester", "127.0.0.1")
//	if err != nil {
//		t.Fatalf("error creating memberlist: %v", err)
//	}
//
//	gcpList, err := gcpMemberList.List()
//	if err != nil {
//		t.Fatalf("error creating memberlist: %v", err)
//	}
//
//	//fmt.Printf("%v\n", gcpList)
//
//	var id string
//	var ipAdress *gcp_memberlist.IpAddress
//
//	var ids []string
//
//	i := 0
//	for k, v := range gcpList {
//
//		id = k
//		ipAdress = v
//
//		ids = append(ids, id)
//
//		message := &chatgroup.Message{
//			MsgType: chatgroup.Message_TEST_PUBLISH_REQUEST,
//			Sender: &chatgroup.Member{
//				Name: testname,
//			},
//			Text: "\\gcp " + id,
//		}
//
//		err = sendMessage(message, ipAdress.Ip+":"+ipAdress.Port)
//		if err != nil {
//			t.Errorf("Failed send reply: %v", err)
//		}
//		i++
//	}
//	//fmt.Printf("%d tests sent\n", i)
//
//	j := 0
//	for {
//		select {
//		case testResult := <-testPublishReplyChannel:
//			//fmt.Printf("Check the test result: %q\n", testResult.Text)
//			ok := false
//			for _, v := range ids {
//				if testResult.Text == v {
//					ok = true
//					fmt.Printf("Result %q as expected %q\n", testResult.Text, v)
//					j++
//					break
//				}
//			}
//			if !ok {
//				t.Errorf("result %q not found\n", testResult.Text)
//			}
//			//fmt.Printf("%d tests checked\n", j)
//			if i == j {
//				return
//			}
//		}
//	}
//}

/*


╰─[:)] % go test -v
=== RUN   TestPingChats
--- FAIL: TestPingChats (10.43s)
main_test.go:128: error creating memberlist: failed to list memberlist (ServiceUrl: https://europe-west1-gke-serverless-211907.cloudfunctions.net
Uuid: "7ae6b3e2-23f4-4250-af81-ad1f88598db7"
Self: 	 { Name: "tester" Ip: "127.0.0.1" Port: "" Protocol: "tcp" } ): Post https://europe-west1-gke-serverless-211907.cloudfunctions.net/list: net/http: TLS handshake timeout
=== RUN   TestQuitChats
--- PASS: TestQuitChats (0.41s)
FAIL
exit status 1
FAIL	bitbucket.org/stefanhans/go-thesis/6.5./rudimentary-chat-tcp	10.858s

*/
