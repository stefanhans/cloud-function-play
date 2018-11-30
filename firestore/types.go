package memberlist

// IpAddress is the struct for the collection
type IpAddress struct {
	Name     string `firestore:"name,omitempty"`
	Ip       string `firestore:"ip,omitempty"`
	Port     string `firestore:"port,omitempty"`
	Protocol string `firestore:"protocol,omitempty"` // "tcp" or "udp"
}

// Structure of the collection
var collectionName string = "IpAddresses"
