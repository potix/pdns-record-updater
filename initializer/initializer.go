package initializer

import (
	import "github.com/mattn/go-sqlite3"
)

// 
func (i *Initialize) insertZone(zoneName) {

// curl -X POST --data '{"name":"example.org.", "kind": "Native", "masters": [], "nameservers": ["ns1.example.org.", "ns2.example.org."]}' -v -H 'X-API-Key: changeme' http://127.0.0.1:8081/api/v1/servers/localhost/zones | jq .

}

func (i *Initialize) insertRecord() {


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


}



// initalize is initialize power dns record
func (i *Initialize) Initialize() (error) {
	for {
		result, err := i.client.GetWatchResult()
		if (err != nil) {
			plog.Error("can not get watcher result (%v)", err) 
			continue;
		}
		time.Sleep(time.Second)
		break
	}

	for zoneName, zoneResult  in range result.zone {
		//INSERT INTO domains (name, type) VALUES ('zoneName', 'NATIVE');

	}

	// initialize

	//INSERT INTO domains (name, type) VALUES ('powerdns.com', 'NATIVE');

}


// New is create initializer
func New(initializer *contexter.Initializer, client client.Client) (*Initializer) {
        return &Initializer {
                client:     client,
		pdnsSqlite: initializer.PdnsSqlite,
        }
}

