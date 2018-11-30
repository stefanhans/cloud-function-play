package memberlist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestReset(t *testing.T) {

	serviceUrl := "https://europe-west1-gke-serverless-211907.cloudfunctions.net"
	contentType := "application/x-www-form-urlencoded"

	var memberlist1, memberlist2 map[string]*IpAddress

	res, err := http.Post(serviceUrl+"/subscribe",
		contentType,
		strings.NewReader(fmt.Sprintf("%s %s %s %s %s",
			"uuid", "name", "ip", "port", "protocol")))
	if err != nil {
		t.Errorf("failed to subscribe memberlist: %v\n", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("failed to read response of memberlist subscription: %v\n", err)
	}

	err = json.Unmarshal(body, &memberlist1)
	if err != nil {
		t.Errorf("failed to unmarshall response of memberlist subscription: %v\n", err)
	}

	if len(memberlist1) != 1 {
		t.Errorf("unexpected number (%d) of members in result: %v\n", 1, len(memberlist1))
	}

	res, err = http.Post(serviceUrl+"/reset",
		contentType, nil)
	res.Body.Close()
	if err != nil {
		t.Errorf("failed to reset memberlist: %v\n", err)
	}

	res, err = http.Post(serviceUrl+"/list",
		contentType, nil)
	if err != nil {
		t.Errorf("failed to list memberlist: %v\n", err)
	}

	body, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("failed to read response of memberlist list: %v\n", err)
	}

	json.Unmarshal(body, &memberlist2)
	if err != nil {
		t.Errorf("failed to unmarshall response of memberlist list: %v\n", err)
	}

	if len(memberlist2) != 0 {
		t.Errorf("unexpected number (%d) of members in result: %v\n", 0, len(memberlist2))
	}
}
