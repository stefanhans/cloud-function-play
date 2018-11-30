package main

import (
	"fmt"
	"os"

	gcp_memberlist "bitbucket.org/stefanhans/go-thesis/Libraries/memberlist"
)

var (
	gcpMemberList *gcp_memberlist.Memberlist
)

// CreateMemberlist creates a memberlist regarding GCP Cloud Functions with Firestore
func CreateMemberlist(name, ip string) (*gcp_memberlist.Memberlist, error) {

	//Get URL for GCP service
	serviceUrl := os.Getenv("GCP_SERVICE_URL")
	if serviceUrl == "" {
		return nil, fmt.Errorf("GCP_SERVICE_URL environment variable unset or missing")
	}

	return gcp_memberlist.Create(
		&gcp_memberlist.IpAddress{
			Name:     name,
			Ip:       ip,
			Port:     "",
			Protocol: "tcp",
		})
}
