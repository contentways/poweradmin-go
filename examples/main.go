// quickstart is a minimal CLI demonstrating the poweradmin-go SDK.
//
// Configure via env:
//
//	POWERADMIN_URL       https://dns.example.com
//	POWERADMIN_API_KEY   your-api-key
//	POWERADMIN_DEBUG     set to 1 to print HTTP requests to stderr
//
// Commands:
//
//	list-zones
//	create-zone <name> [type]                       (default type: MASTER)
//	delete-zone <name>
//	list-records <zone-name>
//	create-record <zone-name> <record-name> <type> <content> [ttl]
//	delete-record <zone-name> <record-id>
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/contentways/poweradmin-go/v4/poweradmin"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	url := mustEnv("POWERADMIN_URL")
	key := mustEnv("POWERADMIN_API_KEY")

	opts := []poweradmin.Option{
		poweradmin.WithBaseURL(url),
		poweradmin.WithAPIKey(key),
		poweradmin.WithRetry(3),
	}
	if os.Getenv("POWERADMIN_DEBUG") == "1" {
		opts = append(opts, poweradmin.WithDebugWriter(os.Stderr))
	}

	client, err := poweradmin.NewClient(opts...)
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()
	cmd, args := os.Args[1], os.Args[2:]

	switch cmd {
	case "list-zones":
		listZones(ctx, client)
	case "create-zone":
		if len(args) < 1 {
			log.Fatal("usage: create-zone <name> [type]")
		}
		zoneType := poweradmin.ZoneTypeMaster
		if len(args) >= 2 {
			zoneType = poweradmin.ZoneType(args[1])
		}
		createZone(ctx, client, args[0], zoneType)
	case "delete-zone":
		if len(args) < 1 {
			log.Fatal("usage: delete-zone <name>")
		}
		deleteZone(ctx, client, args[0])
	case "list-records":
		if len(args) < 1 {
			log.Fatal("usage: list-records <zone-name>")
		}
		listRecords(ctx, client, args[0])
	case "create-record":
		if len(args) < 4 {
			log.Fatal("usage: create-record <zone-name> <record-name> <type> <content> [ttl]")
		}
		ttl := 3600
		if len(args) >= 5 {
			v, err := strconv.Atoi(args[4])
			if err != nil {
				log.Fatalf("invalid ttl: %v", err)
			}
			ttl = v
		}
		createRecord(ctx, client, args[0], args[1], args[2], args[3], ttl)
	case "delete-record":
		if len(args) < 2 {
			log.Fatal("usage: delete-record <zone-name> <record-id>")
		}
		deleteRecord(ctx, client, args[0], args[1])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`quickstart — poweradmin-go sample

Commands:
  list-zones
  create-zone <name> [type]                      (default type: MASTER)
  delete-zone <name>
  list-records <zone-name>
  create-record <zone-name> <name> <type> <content> [ttl]
  delete-record <zone-name> <record-id>

Env: POWERADMIN_URL, POWERADMIN_API_KEY  (POWERADMIN_DEBUG=1 for HTTP logs)`)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("environment variable %s is required", name)
	}
	return v
}

func listZones(ctx context.Context, c *poweradmin.Client) {
	zones, err := c.Zone.All(ctx)
	if err != nil {
		log.Fatalf("list zones: %v", err)
	}
	fmt.Printf("%d zone(s)\n", len(zones))
	for _, z := range zones {
		fmt.Printf("  [%5d] %-40s %s\n", z.ID, z.Name, z.Type)
	}
}

func createZone(ctx context.Context, c *poweradmin.Client, name string, t poweradmin.ZoneType) {
	id, _, err := c.Zone.Create(ctx, poweradmin.ZoneCreateOpts{
		Name: name,
		Type: t,
	})
	if err != nil {
		log.Fatalf("create zone: %v", err)
	}
	fmt.Printf("created zone %s (id %d)\n", name, id)
}

func deleteZone(ctx context.Context, c *poweradmin.Client, name string) {
	z, _, err := c.Zone.GetByName(ctx, name)
	if err != nil {
		log.Fatalf("resolve zone: %v", err)
	}
	if _, err := c.Zone.Delete(ctx, z.ID); err != nil {
		log.Fatalf("delete zone: %v", err)
	}
	fmt.Printf("deleted zone %s (id %d)\n", name, z.ID)
}

func listRecords(ctx context.Context, c *poweradmin.Client, zoneName string) {
	z, _, err := c.Zone.GetByName(ctx, zoneName)
	if err != nil {
		log.Fatalf("resolve zone: %v", err)
	}
	records, err := c.Record.All(ctx, z.ID)
	if err != nil {
		log.Fatalf("list records: %v", err)
	}
	fmt.Printf("%d record(s) in %s\n", len(records), zoneName)
	for _, r := range records {
		fmt.Printf("  [%5s] %-30s %-6s %-30s TTL=%d\n", r.ID, r.Name, r.Type, r.Content, r.TTL)
	}
}

func createRecord(ctx context.Context, c *poweradmin.Client, zoneName, name, recordType, content string, ttl int) {
	z, _, err := c.Zone.GetByName(ctx, zoneName)
	if err != nil {
		log.Fatalf("resolve zone: %v", err)
	}
	id, _, err := c.Record.Create(ctx, z.ID, poweradmin.RecordCreateOpts{
		Name:    name,
		Type:    recordType,
		Content: content,
		TTL:     ttl,
	})
	if err != nil {
		log.Fatalf("create record: %v", err)
	}
	fmt.Printf("created record %s %s %s in zone %s (id %s)\n", name, recordType, content, zoneName, id)
}

func deleteRecord(ctx context.Context, c *poweradmin.Client, zoneName string, recordID string) {
	z, _, err := c.Zone.GetByName(ctx, zoneName)
	if err != nil {
		log.Fatalf("resolve zone: %v", err)
	}
	if _, err := c.Record.Delete(ctx, z.ID, recordID); err != nil {
		log.Fatalf("delete record: %v", err)
	}
	fmt.Printf("deleted record %s from zone %s\n", recordID, zoneName)
}
