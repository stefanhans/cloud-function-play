package memberlist

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Subscribe get information of the new member via http and stores it in Firestore
func Subscribe(w http.ResponseWriter, r *http.Request) {

	// Get rid of warnings
	_ = r

	// Sets your Google Cloud Platform project ID.
	projectId := os.Getenv("GCP_PROJECT")
	if projectId == "" {
		http.Error(w, fmt.Sprintf("GCP_PROJECT environment variable unset or missing"), http.StatusInternalServerError)
	}

	// Get a Firestore client.
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create client: %v", err), http.StatusInternalServerError)
	}
	defer client.Close()

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request: %s", err), http.StatusInternalServerError)
		panic(err)
	}

	// Create a struct out of the body (except first word, i.e. UUID as key/document)
	ipAddress := &IpAddress{
		Name:     strings.Split(string(body), " ")[1],
		Ip:       strings.Split(string(body), " ")[2],
		Port:     strings.Split(string(body), " ")[3],
		Protocol: strings.Split(string(body), " ")[4],
	}

	// Marshall the struct to JSON
	ipAddressJson, err := json.MarshalIndent(ipAddress, "", "  ")
	if err != nil {
		// Return HTTP error code 500 Internal Server Error
		http.Error(w, fmt.Sprintf("failed to marshall %v: %s", ipAddress, err), http.StatusInternalServerError)
	}

	// Save the JSON as string in filed named by the UUID and as document named by the same UUID
	_, err = client.Collection(collectionName).Doc(strings.Split(string(body), " ")[0]).Set(ctx, map[string]interface{}{
		strings.Split(string(body), " ")[0]: fmt.Sprintf("%v", string(ipAddressJson)),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to subscribe %v: %s", ipAddress, err), http.StatusInternalServerError)
	}

	// Get iterator over the documents and members, respectively
	iter := client.Collection(collectionName).Documents(ctx)

	// Map of IpAddress
	var ipAdresses map[string]IpAddress = make(map[string]IpAddress)

	// Iterate over the documents
	var ipAddr IpAddress
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to iterate over collection %q: %s", collectionName, err), http.StatusInternalServerError)
		}

		// Get the JSON string, unmarshall it, and insert it into the map
		if v, ok := doc.Data()[doc.Ref.ID].(string); ok {
			json.Unmarshal([]byte(v), &ipAddr)
		} else {
			http.Error(w, fmt.Sprintf("failed to convert %q: %v", collectionName, err), http.StatusInternalServerError)
		}
		ipAdresses[doc.Ref.ID] = ipAddr
	}

	// Marshall all and return it as response
	ipAddressesJson, err := json.Marshal(ipAdresses)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshall %q: %v", collectionName, err), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "%v", string(ipAddressesJson))
}

// DO NOT FORGET:
// $ export GCP_PROJECT="gke-serverless-211907"
// $ export GOOGLE_APPLICATION_CREDENTIALS="/Users/stefan/.secret/gke-serverless-211907-181ed186fa7f.json"

// cd ~/go/src/bitbucket.org/stefanhans/go-thesis/6.5./cloud-functions
// gcloud alpha functions deploy subscribe --region europe-west1 --entry-point Subscribe --runtime go111 --trigger-http

// curl -d "uuid memberlist 127.0.0.1 22365 tcp" https://europe-west1-gke-serverless-211907.cloudfunctions.net/subscribe
