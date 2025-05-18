package main

import (
	"log"

	"github.com/gocql/gocql"
)

func main() {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "my_app"
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Insert example
	if err := session.Query(`INSERT INTO users (id, name, email) VALUES (?, ?, ?)`,
		gocql.TimeUUID(), "John Doe", "john@example.com").Exec(); err != nil {
		log.Fatal(err)
	}

}
