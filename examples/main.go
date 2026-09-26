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
//
// DNSSEC and server status (Poweradmin 4.5+):
//
//	dnssec-status <zone-name>
//	dnssec-enable <zone-name>
//	dnssec-disable <zone-name>
//	list-keys <zone-name>
//	add-key <zone-name> <ksk|zsk|csk> <algorithm> <bits> (e.g. csk ecdsa256 256)
//	activate-key <zone-name> <key-id>
//	deactivate-key <zone-name> <key-id>
//	delete-key <zone-name> <key-id>
//	rectify <zone-name>
//	server-status [metric,...]                     (e.g. uptime,udp-queries)
package main

import (
	"context"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

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
	case "dnssec-status":
		requireArgs(args, 1, "dnssec-status <zone-name>")
		dnssecStatus(ctx, client, args[0])
	case "dnssec-enable", "dnssec-disable":
		requireArgs(args, 1, cmd+" <zone-name>")
		setDNSSEC(ctx, client, args[0], cmd == "dnssec-enable")
	case "list-keys":
		requireArgs(args, 1, "list-keys <zone-name>")
		listKeys(ctx, client, args[0])
	case "add-key":
		requireArgs(args, 4, "add-key <zone-name> <ksk|zsk|csk> <algorithm> <bits>")
		addKey(ctx, client, args[0], poweradmin.DNSSECKeyType(args[1]), args[2], mustInt(args[3], "bits"))
	case "activate-key", "deactivate-key":
		requireArgs(args, 2, cmd+" <zone-name> <key-id>")
		setKeyActive(ctx, client, args[0], mustInt(args[1], "key id"), cmd == "activate-key")
	case "delete-key":
		requireArgs(args, 2, "delete-key <zone-name> <key-id>")
		deleteKey(ctx, client, args[0], mustInt(args[1], "key id"))
	case "rectify":
		requireArgs(args, 1, "rectify <zone-name>")
		rectify(ctx, client, args[0])
	case "server-status":
		var metrics []string
		if len(args) >= 1 {
			metrics = strings.Split(args[0], ",")
		}
		serverStatus(ctx, client, metrics)
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

DNSSEC and server status (Poweradmin 4.5+):
  dnssec-status <zone-name>
  dnssec-enable <zone-name>
  dnssec-disable <zone-name>
  list-keys <zone-name>
  add-key <zone-name> <ksk|zsk|csk> <algorithm> <bits>   (e.g. csk ecdsa256 256)
  activate-key <zone-name> <key-id>
  deactivate-key <zone-name> <key-id>
  delete-key <zone-name> <key-id>
  rectify <zone-name>
  server-status [metric,...]                     (e.g. uptime,udp-queries)

Env: POWERADMIN_URL, POWERADMIN_API_KEY  (POWERADMIN_DEBUG=1 for HTTP logs)`)
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("environment variable %s is required", name)
	}
	return v
}

func requireArgs(args []string, n int, usage string) {
	if len(args) < n {
		log.Fatalf("usage: %s", usage)
	}
}

func mustInt(s, what string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid %s %q: %v", what, s, err)
	}
	return v
}

func mustZoneID(ctx context.Context, c *poweradmin.Client, name string) int {
	z, _, err := c.Zone.GetByName(ctx, name)
	if err != nil {
		log.Fatalf("resolve zone: %v", err)
	}
	return z.ID
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

func printDNSSEC(zoneName string, d *poweradmin.ZoneDNSSEC) {
	state := "unsigned"
	if d.Enabled {
		state = "signed"
	}
	if d.Presigned {
		state += " (presigned, managed at the primary)"
	}
	fmt.Printf("%s: %s\n", zoneName, state)
	for _, ds := range d.DSRecords {
		fmt.Printf("  DS %d %d %d %s\n", ds.KeyTag, ds.Algorithm, ds.DigestType, ds.Digest)
	}
}

func dnssecStatus(ctx context.Context, c *poweradmin.Client, zoneName string) {
	d, _, err := c.Zone.GetDNSSEC(ctx, mustZoneID(ctx, c, zoneName))
	if err != nil {
		log.Fatalf("get dnssec: %v", err)
	}
	printDNSSEC(zoneName, d)
}

func setDNSSEC(ctx context.Context, c *poweradmin.Client, zoneName string, enabled bool) {
	d, _, err := c.Zone.SetDNSSEC(ctx, mustZoneID(ctx, c, zoneName), enabled)
	if err != nil {
		log.Fatalf("set dnssec: %v", err)
	}
	printDNSSEC(zoneName, d)
}

func printKey(k *poweradmin.DNSSECKey) {
	algorithm := k.Algorithm
	if algorithm == "" {
		algorithm = fmt.Sprintf("alg-%d", k.AlgorithmID)
	}
	active := "inactive"
	if k.Active {
		active = "active"
	}
	fmt.Printf("  [%3d] %-3s tag=%-5d %-18s %4d bits  %s\n", k.ID, strings.ToUpper(string(k.Type)), k.KeyTag, algorithm, k.Bits, active)
}

func listKeys(ctx context.Context, c *poweradmin.Client, zoneName string) {
	keys, _, err := c.DNSSEC.ListKeys(ctx, mustZoneID(ctx, c, zoneName))
	if err != nil {
		log.Fatalf("list keys: %v", err)
	}
	fmt.Printf("%d key(s) in %s\n", len(keys), zoneName)
	for _, k := range keys {
		printKey(k)
	}
}

func addKey(ctx context.Context, c *poweradmin.Client, zoneName string, keyType poweradmin.DNSSECKeyType, algorithm string, bits int) {
	k, _, err := c.DNSSEC.AddKey(ctx, mustZoneID(ctx, c, zoneName), poweradmin.DNSSECKeyCreateOpts{
		Type:      keyType,
		Algorithm: algorithm,
		Bits:      bits,
	})
	if err != nil {
		log.Fatalf("add key: %v", err)
	}
	fmt.Printf("added key to %s (PowerDNS creates keys inactive; use activate-key):\n", zoneName)
	printKey(k)
}

func setKeyActive(ctx context.Context, c *poweradmin.Client, zoneName string, keyID int, active bool) {
	k, _, err := c.DNSSEC.SetKeyActive(ctx, mustZoneID(ctx, c, zoneName), keyID, active)
	if err != nil {
		log.Fatalf("update key: %v", err)
	}
	printKey(k)
}

func deleteKey(ctx context.Context, c *poweradmin.Client, zoneName string, keyID int) {
	if _, err := c.DNSSEC.DeleteKey(ctx, mustZoneID(ctx, c, zoneName), keyID); err != nil {
		log.Fatalf("delete key: %v", err)
	}
	fmt.Printf("deleted key %d from zone %s\n", keyID, zoneName)
}

func rectify(ctx context.Context, c *poweradmin.Client, zoneName string) {
	if _, err := c.DNSSEC.Rectify(ctx, mustZoneID(ctx, c, zoneName)); err != nil {
		log.Fatalf("rectify: %v", err)
	}
	fmt.Printf("rectified zone %s\n", zoneName)
}

func serverStatus(ctx context.Context, c *poweradmin.Client, metrics []string) {
	s, _, err := c.Server.Status(ctx, poweradmin.ServerStatusOpts{Metrics: metrics})
	if poweradmin.IsServiceUnavailable(err) {
		fmt.Println("PowerDNS is not reachable")
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("server status: %v", err)
	}

	uptime := "unknown"
	if s.UptimeSeconds != nil {
		uptime = fmt.Sprintf("%ds", *s.UptimeSeconds)
	}
	fmt.Printf("PowerDNS %s %s (server %s), uptime %s\n", s.DaemonType, s.Version, s.ServerID, uptime)

	for _, name := range slices.Sorted(maps.Keys(s.Metrics)) {
		fmt.Printf("  %-30s %s\n", name, s.Metrics[name])
	}
}
