package main

// Service is the struct for the collection
type Service struct {
	Name        string `firestore:"name,omitempty"`
	Url         string `firestore:"url,omitempty"`
	Description string `firestore:"description,omitempty"`
}
