package memberlist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/satori/go.uuid"
)

// todo: handling html errors (e.g. Unsubscribe())

// Memberlist is the core struct for all relevant data
type Memberlist struct {
	ServiceUrl string
	Uuid       string
	Self       *IpAddress
}

func (ml *Memberlist) String() string {

	out := fmt.Sprintf("ServiceUrl: %s\n", ml.ServiceUrl)
	out += fmt.Sprintf("Uuid: %q\n", ml.Uuid)
	out += fmt.Sprintf("Self: %v", strings.TrimRight((*ml.Self).String(), "\n"))

	return out
}

// Create returns the memberlist for oneself
func Create(self *IpAddress) (*Memberlist, error) {

	serviceUrl := os.Getenv("GCP_SERVICE_URL")
	if serviceUrl == "" {
		return nil, fmt.Errorf("GCP_SERVICE_URL environment variable unset or missing")
	}

	id, _ := uuid.NewV4()

	ipAddresses := make(map[string]*IpAddress)
	ipAddresses[id.String()] = self

	return &Memberlist{
		ServiceUrl: serviceUrl,
		Uuid:       id.String(),
		Self:       self,
	}, nil
}

// IpAddress is the struct for the Firestore
type IpAddress struct {
	Name     string `firestore:"name,omitempty"`
	Ip       string `firestore:"ip,omitempty"`
	Port     string `firestore:"port,omitempty"`
	Protocol string `firestore:"protocol,omitempty"` // "tcp" or "udp"
}

func (ia *IpAddress) String() string {

	out := fmt.Sprintf("\t { Name: %q", ia.Name)
	out += fmt.Sprintf(" Ip: %q", ia.Ip)
	out += fmt.Sprintf(" Port: %q", ia.Port)
	out += fmt.Sprintf(" Protocol: %q } \n", ia.Protocol)

	return out
}

// Subscribe stores a new member and returns the current memberlist
func (ml *Memberlist) Subscribe() (map[string]*IpAddress, error) {

	// Send request to service
	res, err := http.Post(ml.ServiceUrl+"/subscribe",
		"application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("%s %s %s %s %s",
			ml.Uuid, ml.Self.Name, ml.Self.Ip, ml.Self.Port, ml.Self.Protocol)))
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe memberlist (%v): %v\n", ml, err)
	}

	// Read response body in JSON
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response of memberlist subscription (%v): %v\n", ml, err)
	}

	// Unmarshall all ip addresses as map
	var memberlist map[string]*IpAddress
	err = json.Unmarshal(body, &memberlist)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshall response of memberlist subscription (%v): %v\n", res.Proto, err)
	}

	return memberlist, nil
}

// Unsubscribe removes oneself from the memberlist
func (ml *Memberlist) Unsubscribe() error {

	// Send request to service
	res, err := http.Post(ml.ServiceUrl+"/unsubscribe",
		"application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("%s", ml.Uuid)))
	if err != nil {
		return fmt.Errorf("failed to unsubscribe memberlist (%v): %v\n", (*ml), err)
	}

	// Read response body in JSON
	_, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

// List returns the memberlist from Firestore
func (ml *Memberlist) List() (map[string]*IpAddress, error) {

	// Send request to service
	res, err := http.Post(ml.ServiceUrl+"/list",
		"application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberlist (%v): %v\n", ml, err)
	}

	// Read response body in JSON
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response of memberlist list (%v): %v\n", ml, err)
	}

	// Unmarshall all ip addresses as map
	var memberlist map[string]*IpAddress
	json.Unmarshal(body, &memberlist)

	return memberlist, nil
}

// Reset clears the memberlist on Firestore
func Reset(serviceUrl string) error {

	// Send request to service
	res, err := http.Post(serviceUrl+"/reset",
		"application/x-www-form-urlencoded", nil)
	res.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to reset memberlist (%v): %v\n", serviceUrl, err)
	}

	return nil
}
