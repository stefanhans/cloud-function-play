package memberlist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// List returns the list of members from Firestore
func List(w http.ResponseWriter, r *http.Request) {

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
// gcloud alpha functions deploy list --region europe-west1 --entry-point List --runtime go111 --trigger-http

// curl https://europe-west1-gke-serverless-211907.cloudfunctions.net/list
