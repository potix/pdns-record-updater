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
	"time"
	"strings"
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
	stmt, err := db.Prepare( `INSERT INTO "domains" ("name", "type", "account") VALUES (?, ?, ?)`);
	if err != nil {
		return 0, errors.Wrap(err, "can not prepare of domain")
	}
	result, err := stmt.Exec(domain, "NATIVE", strings.Replace(zoneWatchResultResponse.NameServer[0].Email, "@", ".", -1))
	if err != nil {
		return 0, errors.Wrap(err, "can not execute statement of domain")
	}
	domainId, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "can not get domain id")
	}
	stmt.Close()

	return domainId, nil
}

func (i *Initializer) insertRecord(db *sql.DB, domainId int64, domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	stmt, err := db.Prepare( `INSERT INTO "records" ("domain_id", "name", "type", "content", "ttl", "prio", disable, auth) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return errors.Wrap(err, "can not prepare of domain")
	}
	// ns record
	for _, nameserver := range zoneWatchResultResponse.NameServer {
		if nameserver.Type != "A" && nameserver.Type != "AAAA" {
			continue
		}
		_, err = stmt.Exec(domainId, domain, "NS", nameserver.Name, nameserver.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of ns record")
		}
	}
	// name server record
	for _, nameserver := range zoneWatchResultResponse.NameServer {
		_, err = stmt.Exec(domainId, nameserver.Name, nameserver.Type, nameserver.Content, nameserver.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of name server record")
		}
	}
	// static record
	for _, staticRecord := range zoneWatchResultResponse.StaticRecord {
		_, err = stmt.Exec(domainId, staticRecord.Name, staticRecord.Type, staticRecord.Content, staticRecord.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of static record")
		}
	}
	// dynamic record
	for _, dynamicRecord := range zoneWatchResultResponse.DynamicRecord {
		_, err = stmt.Exec(domainId, dynamicRecord.Name, dynamicRecord.Type, dynamicRecord.Content, dynamicRecord.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of dynamic record")
		}
	}
	stmt.Close()

	return nil
}

func (i *Initializer) insert(watchResultResponse *structure.WatchResultResponse) (error) {
	db, err := sql.Open("sqlite3", i.initializerContext.PdnsSqlitePath);
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("can not open powedns sqlite (%v)", i.initializerContext.PdnsSqlitePath))
	}
	defer db.Close();
	for domain, zoneWatchResultResponse := range watchResultResponse.Zone {
		domainId, err := i.insertDomain(db, domain, zoneWatchResultResponse)
		if err != nil {
			return err;
		}
		err = i.insertRecord(db, domainId, domain, zoneWatchResultResponse)
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

