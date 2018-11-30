package memberlist

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Reset deletes all documents of the collection from Firestore
func Reset(w http.ResponseWriter, r *http.Request) {

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

	// Iterate over the documents
	iter := client.Collection(collectionName).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to iterate over collection %q: %s", collectionName, err), http.StatusInternalServerError)
		}

		// Delete the document
		doc.Ref.Delete(ctx)
	}
}

// DO NOT FORGET:
// $ export GCP_PROJECT="gke-serverless-211907"
// $ export GOOGLE_APPLICATION_CREDENTIALS="/Users/stefan/.secret/gke-serverless-211907-181ed186fa7f.json"

// cd ~/go/src/bitbucket.org/stefanhans/go-thesis/6.5./cloud-functions
// gcloud alpha functions deploy reset --region europe-west1 --entry-point Reset --runtime go111 --trigger-http

// curl https://europe-west1-gke-serverless-211907.cloudfunctions.net/reset
