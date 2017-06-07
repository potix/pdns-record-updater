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
)

// Initializer is initializer
type Initializer struct {
	client *client.Client
	initializerContext *contexter.Initializer
}

func (i *Initializer) insertRecord(domain string, nameServer []*structure.NameServerRecordWatchResultResponse, staticRecord []*structure.StaticRecordWatchResultResponse, dynamicRecord []*structure.DynamicRecordWatchResultResponse) (error) {

//# Combined replacement of multiple RRsets
//curl -X PATCH --data '{"rrsets": [
//  {"name": "test1.example.org.",
//   "type": "A",
//   "ttl": 86400,
//   "changetype": "REPLACE",
//   "records": [ {"content": "192.0.2.5", "disabled": false} ]
//  },
//  {"name": "test2.example.org.",
//   "type": "AAAA",
//   "ttl": 86400,
//   "changetype": "REPLACE",
//   "records": [ {"content": "2001:db8::6", "disabled": false} ]
//  }
//  ] }' -H 'X-API-Key: changeme' http://127.0.0.1:8081/api/v1/servers/localhost/zones/example.org. | jq .

// INSERT INTO domains (name, type) VALUES (’example.com’, ‘MASTER’);
// INSERT INTO records (domain_id, name, content, type, ttl, prio) VALUES (1, ‘example.com’, ‘ns1.example.com hostmaster.example.com 1′, ‘SOA’, 86400, NULL);
// INSERT INTO records (domain_id, name, content, type, ttl, prio) VALUES (1, ‘example.com’, ‘ns1.example.com’, ‘NS’, 86400, NULL);
// INSERT INTO records (domain_id, name, content, type, ttl, prio) VALUES (1, ‘example.com’, ‘ns2.example.com’, ‘NS’, 86400, NULL);
// INSERT INTO records (domain_id, name, content, type, ttl, prio) VALUES (1, ‘ns1.example.com’, ‘10.0.0.10′, ‘A’, 86400, NULL);
// INSERT INTO records (domain_id, name, content, type, ttl, prio) VALUES (1, ‘ns2.example.com’, ‘10.0.0.20′, ‘A’, 86400, NULL);

	return nil
}



// Initialize is initialize power dns record
func (i *Initializer) Initialize() (err error) {
	var watchResultResponse *structure.WatchResultResponse
	for {
		watchResultResponse, err = i.client.GetWatchResult()
		if (err != nil) {
			belog.Error("can not get watcher result (%v)", err)
			continue;
		}
		time.Sleep(time.Second)
		break
	}
	for domain, zoneWatchResultResponse := range watchResultResponse.Zone {
		// record
		err := i.insertRecord(domain, zoneWatchResultResponse.NameServer, zoneWatchResultResponse.StaticRecord, zoneWatchResultResponse.DynamicRecord)
		if err != nil {
			return err;
		}
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

