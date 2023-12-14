package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	// IMAP server credentials and settings
	username := "email@example.com"
	password := "password"
	server := "ssl0.ovh.net"
	port := "993"
	ssl := true
	excludedDomains := []string{"@internal-domain.org", "@an-other-domain.org"} // Change this to your excluded domains

	// Connect to the IMAP server
	var c *client.Client
	var err error
	if ssl {
		c, err = client.DialTLS(server+":"+port, nil)
	} else {
		c, err = client.Dial(server + ":" + port)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	// Login to the server
	if err := c.Login(username, password); err != nil {
		log.Fatal(err)
	}

	// Open a specific mailbox (folder)
	// Sent, Drafts, Junk, Trash, INBOX, ...
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// Search for all messages in the selected mailbox
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(1, mbox.Messages)

	messages := make(chan *imap.Message, mbox.Messages)
	go func() {
		if err := c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope}, messages); err != nil {
			log.Fatal(err)
		}
	}()

	// Create a CSV file to save contacts
	file, err := os.Create("contacts.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	if err := writer.Write([]string{"Name", "Email"}); err != nil {
		log.Fatal(err)
	}

	// Map to track processed email addresses
	processedEmails := make(map[string]bool)

	// Parse contacts and save them in the CSV file, excluding specific domain emails
	for msg := range messages {
		for _, addr := range msg.Envelope.To {
			email := addr.Address()
			if !processedEmails[email] && !containsExcludedDomain(email, excludedDomains) {
				if err := writer.Write([]string{addr.PersonalName, email}); err != nil {
					log.Fatal(err)
				}
				processedEmails[email] = true
			}
		}
	}

	log.Println("Contacts saved to contacts.csv")
}

// Function to check if an email contains any excluded domain
func containsExcludedDomain(email string, excludedDomains []string) bool {
	for _, domain := range excludedDomains {
		if strings.Contains(email, domain) {
			return true
		}
	}
	return false
}