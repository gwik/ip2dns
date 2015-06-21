package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/route53"
	"github.com/nightlyone/lockfile"
)

var (
	host   string
	zoneID string

	lfPath = path.Join(os.TempDir(), "dynip2route53.lock")

	awsAccessKey string
	awsSecretKey string
)

var client = new(http.Client)

func getIP(client *http.Client) (net.IP, error) {
	resp, err := client.Get("https://ip.appspot.com")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}
	s, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(strings.Trim(s, " \n"))
	if ip == nil {
		return nil, fmt.Errorf("Failed to parse ip from: `%v`", s)
	}
	return ip, err
}

func change(c *route53.Route53, ip net.IP, record route53.ResourceRecordSet) error {

	req := route53.ChangeResourceRecordSetsRequest{
		Comment: "changed by ip2dns",
		Changes: []route53.Change{
			route53.Change{
				Action: "UPSERT",
				Record: route53.ResourceRecordSet{
					Name:    record.Name,
					Type:    record.Type,
					TTL:     record.TTL,
					Records: []string{ip.String()},
				},
			},
		},
	}

	// log.Printf("change req: %#v", req)
	// resp, err := c.ChangeResourceRecordSets(zoneID, &req)
	_, err := c.ChangeResourceRecordSets(zoneID, &req)
	if err != nil {
		return err
	}
	// log.Printf("change: %#v", resp)
	return nil
}

func checkDNS(host string, ip net.IP) (bool, error) {
	addrs, err := net.LookupIP(host)
	if err != nil {
		return false, err
	}

	for _, addr := range addrs {
		if bytes.Equal(addr.To4(), ip) {
			return true, nil
		}
	}
	return false, nil
}

func init() {
	flag.StringVar(&host, "host", "", "[REQUIRED] Host name to set.")
	flag.StringVar(&zoneID, "zoneid", "", "[REQUIRED] AWS Route53 zone ID.")
	flag.StringVar(&awsAccessKey, "accesskey", "", "AWS access key")
	flag.StringVar(&awsSecretKey, "secretkey", "", "AWS secret key")
	flag.StringVar(&lfPath, "lock", lfPath, "Path for the lock file.")
}

func main() {

	flag.Parse()

	if host == "" {
		fmt.Fprintln(os.Stderr, "host required.")
		os.Exit(1)
	}

	if !strings.HasSuffix(host, ".") {
		host = host + "."
	}

	if zoneID == "" {
		fmt.Fprintln(os.Stderr, "zoneid required.")
		os.Exit(1)
	}

	lf, err := lockfile.New(lfPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := lf.TryLock(); err != nil {
		log.Fatal(err)
	}
	defer lf.Unlock()

	log.Printf("lock at: %s", lfPath)

	ip, err := getIP(client)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Your ip: %s", ip)
	ip = ip.To4()
	if ip == nil {
		log.Fatal("Not a IPv4 address: %s", ip)
	}

	// avoid querying the API if value is already correct.
	ok, err := checkDNS(host, ip)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		log.Println("DNS Lookup OK, nothing to change.")
		return
	}

	auth, err := aws.GetAuth(awsAccessKey, awsSecretKey)
	if err != nil {
		log.Fatal(err)
	}
	r := route53.New(auth, aws.EUWest)

	rrs, err := r.ListResourceRecordSets(zoneID, &route53.ListOpts{
		Name: host,
		Type: "A",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, rec := range rrs.Records {
		if rec.Name == host {
			if len(rec.Records) != 1 {
				log.Fatal("Record not set.")
			}
			curIP := rec.Records[0]
			if curIP != ip.String() {
				err := change(r, ip, rec)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Println("Nothing to change.")
			}
			return
		}
	}
	log.Fatal("Record not found.")
}
