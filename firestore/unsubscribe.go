package memberlist

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
)

// Unsubscribe deletes the member specified by UUID in Firestore
func Unsubscribe(w http.ResponseWriter, r *http.Request) {

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

	// Delete document specified by UUID
	_, err = client.Collection(collectionName).Doc(strings.Split(string(body), " ")[0]).Delete(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete %q: %s", strings.Split(string(body), " ")[0], err), http.StatusInternalServerError)
	}
}

// DO NOT FORGET:
// $ export GCP_PROJECT="gke-serverless-211907"
// $ export GOOGLE_APPLICATION_CREDENTIALS="/Users/stefan/.secret/gke-serverless-211907-181ed186fa7f.json"

// cd ~/go/src/bitbucket.org/stefanhans/go-thesis/6.5./cloud-functions
// gcloud alpha functions deploy unsubscribe --region europe-west1 --entry-point Unsubscribe --runtime go111 --trigger-http

// curl -d "uuid" https://europe-west1-gke-serverless-211907.cloudfunctions.net/unsubscribe
