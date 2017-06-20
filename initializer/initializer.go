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

func (i *Initializer) selectDomain(db *sql.DB, domain string) (bool, error) {
	rows, err := db.Query("SELECT * FROM domains WHERE name = ?", domain)
	defer rows.Close()
	if err != nil {
		return false, err
	}
	for rows.Next() {
		return true, nil
	}
	return false, nil
}

func (i *Initializer) insertDomain(db *sql.DB, domain string) (int64, error) {
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

	return domainID, nil
}

func (i *Initializer) insertRecord(db *sql.DB, domainID int64, domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	stmt, err := db.Prepare(`INSERT INTO "records" ("domain_id", "name", "type", "content", "ttl", "prio", "disabled", "auth") VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return errors.Wrap(err, "can not prepare of domain")
	}
	defer stmt.Close()
	// soa record
	soaMinimumTtl := i.initializerContext.SoaMinimumTTL
	if soaMinimumTtl == 0 {
		soaMinimumTtl = 60
	}
	content := fmt.Sprintf("%v %v 1 10800 3600 604800 %v", helper.DotHostname(zoneWatchResultResponse.PrimaryNameServer, domain), helper.DotEmail(zoneWatchResultResponse.Email), soaMinimumTtl)
	_, err = stmt.Exec(domainID, helper.NoDotDomain(domain), "SOA", content, 3600, 0, 0, 1);
	if err != nil {
		return errors.Wrap(err, "can not execute statement of soa record")
	}
	// ns record
	for _, nameServer := range zoneWatchResultResponse.NameServerList {
		if nameServer.Type != "A" && nameServer.Type != "AAAA" {
			continue
		}
		_, err = stmt.Exec(domainID, helper.NoDotDomain(domain), "NS", helper.DotHostname(nameServer.Name, domain), nameServer.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of ns record")
		}
	}
	// name server record
	for _, nameServer := range zoneWatchResultResponse.NameServerList {
		name := helper.FixupRrsetName(nameServer.Name, domain, nameServer.Type, false)
		content := helper.FixupRrsetContent(nameServer.Content, domain, nameServer.Type, true)
		_, err = stmt.Exec(domainID, name, nameServer.Type, content, nameServer.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of name server record")
		}
	}
	// static record
	for _, staticRecord := range zoneWatchResultResponse.StaticRecordList {
		name := helper.FixupRrsetName(staticRecord.Name, domain, staticRecord.Type, false)
		content := helper.FixupRrsetContent(staticRecord.Content, domain, staticRecord.Type, true)
		_, err = stmt.Exec(domainID, name, staticRecord.Type, content, staticRecord.TTL, 0, 0, 1);
		if err != nil {
			return errors.Wrap(err, "can not execute statement of static record")
		}
	}
	// dynamic record
	for _, dynamicRecord := range zoneWatchResultResponse.DynamicRecordList {
		name := helper.FixupRrsetName(dynamicRecord.Name, domain, dynamicRecord.Type, false)
		content := helper.FixupRrsetContent(dynamicRecord.Content, domain, dynamicRecord.Type, true)
		_, err = stmt.Exec(domainID, name, dynamicRecord.Type, content, dynamicRecord.TTL, 0, 0, 1);
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
	for domain, zoneWatchResultResponse := range watchResultResponse.ZoneMap {
		exists, err := i.selectDomain(db, domain)
		if err != nil {
			return err;
		}
		if exists {
			belog.Info("domain is already exists")
			continue
		}
		domainID, err := i.insertDomain(db, domain)
		if err != nil {
			return errors.Wrap(err, "can not insert domain");
		}
		err = i.insertRecord(db, domainID, domain, zoneWatchResultResponse)
		if err != nil {
			return errors.Wrap(err, "can not insert record");
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
			time.Sleep(time.Second)
			continue;
		}
		time.Sleep(time.Second)
		break
	}
	if err = i.insert(watchResultResponse); err != nil {
		return errors.Wrap(err, "can not initialize");
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

