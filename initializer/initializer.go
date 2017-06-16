package initializer

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
        "database/sql"
	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
        "github.com/potix/pdns-record-updater/api/structure"
        "github.com/potix/pdns-record-updater/helper"
	"time"
	"fmt"
)

// Initializer is initializer
type Initializer struct {
	client *client.Client
	initializerContext *contexter.Initializer
}

func (i *Initializer) insertDomain(db *sql.DB, domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (int64, error) {
	if len(zoneWatchResultResponse.NameServer) == 0 {
		return 0, errors.Errorf("not name server")
	}
	stmt, err := db.Prepare( `INSERT INTO "domains" ("name", "type") VALUES (?, ?)`);
	if err != nil {
		return 0, errors.Wrap(err, "can not prepare of domain")
	}
	defer stmt.Close()
	result, err := stmt.Exec(helper.NoDotDomain(domain), "NATIVE")
	if err != nil {
		return 0, errors.Wrap(err, "can not execute statement of domain")
	}
	domainID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "can not get domain id")
	}

	return domainId, nil
}

func (i *Initializer) insertRecord(db *sql.DB, domainID int64, domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	stmt, err := db.Prepare(`INSERT INTO "records" ("domain_id", "name", "type", "content", "ttl", "prio", disable, auth) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return errors.Wrap(err, "can not prepare of domain")
	}
	defer stmt.Close()
	// soa record
	var primary *structure.NameServerRecordWatchResultResponse
	for _, nameServer := range zoneWatchResultResponse.NameServer {
		if nameServer.Type != "A" && nameServer.Type != "AAAA" {
			continue
		}
		primary = nameServer
	}
	if primary != nil {
		content := fmt.Printf("%v %v 1 10800 3600 604800 60", helper.DotHostname(primary.Name, domain), helper.DotEmail(primary.Email))
		_, err = stmt.Exec(domainId, helper.NoDotDomain(domain), "SOA", content, primary.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of soa record")
		}
	}
	// ns record
	for _, nameServer := range zoneWatchResultResponse.NameServer {
		if nameserver.Type != "A" && nameserver.Type != "AAAA" {
			continue
		}
		_, err = stmt.Exec(domainId, helper.NoDotDomain(domain), "NS", helper.DotHostname(nameserver.Name, domain), nameserver.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of ns record")
		}
	}
	// name server record
	for _, nameServer := range zoneWatchResultResponse.NameServer {
		name := helper.FixupRrsetName(nameServer.Name, domain, nameServer.Type, false)
		content := helper.FixupRrsetName(nameServer.Content, domain, nameServer.Type, true)
		_, err = stmt.Exec(domainId, name, nameServer.Type, nameServer.Content, nameServer.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of name server record")
		}
	}
	// static record
	for _, staticRecord := range zoneWatchResultResponse.StaticRecord {
		name := helper.FixupRrsetName(staticRecord.Name, domain, staticRecord.Type, false)
		content := helper.FixupRrsetName(staticRecord.Content, domain, staticRecord.Type, true)
		_, err = stmt.Exec(domainId, name, staticRecord.Type, content, staticRecord.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of static record")
		}
	}
	// dynamic record
	for _, dynamicRecord := range zoneWatchResultResponse.DynamicRecord {
		name := helper.FixupRrsetName(dynamicRecord.Name, domain, dynamicRecord.Type, false)
		content := helper.FixupRrsetName(dynamicRecord.Content, domain, dynamicRecord.Type, true)
		_, err = stmt.Exec(domainId, name, dynamicRecord.Type, content, dynamicRecord.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of dynamic record")
		}
	}

	return nil
}

func (i *Initializer) insert(watchResultResponse *structure.WatchResultResponse) (error) {
	db, err := sql.Open("sqlite3", i.initializerContext.PdnsSqlitePath);
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not open powedns sqlite (%v)", i.initializerContext.PdnsSqlitePath))
	}
	defer db.Close();
	for domain, zoneWatchResultResponse := range watchResultResponse.Zone {
		domainID, err := i.insertDomain(db, domain, zoneWatchResultResponse)
		if err != nil {
			return err;
		}
		err = i.insertRecord(db, domainID, domain, zoneWatchResultResponse)
		if err != nil {
			return err;
		}
	}

	return nil
}

// Initialize is initialize power dns record
func (i *Initializer) Initialize() (err error) {
	var watchResultResponse *structure.WatchResultResponse
	for {
		if watchResultResponse, err = i.client.GetWatchResult(); err != nil {
			belog.Error("can not get watcher result (%v)", err)
			continue;
		}
		time.Sleep(time.Second)
		break
	}
	if err = i.insert(watchResultResponse); err != nil {
		return err
	}

	return nil
}

// New is create initializer
func New(initializerContext *contexter.Initializer, client *client.Client) (*Initializer) {
        return &Initializer {
                client:     client,
		initializerContext: initializerContext,
        }
}

