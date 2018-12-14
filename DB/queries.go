package DB

import (
	"strconv"
	"time"
	"github.com/gocql/gocql"
	"fmt"
)

//FileBrowserDBSession is an interface for querying the database session
type FileBrowserDBSession interface {
	GetLatestSnapshotsByTenant(string) ([]string, error)
	GetSnapshotByTenantSerialNumberAndDate(string, string, string) (string, error)
	GetValidTimestampsOfSystem(string, string) ([]time.Time, error)
	GetSystemsOfTenant(string) ([]string, error)
	GetValidTenants() ([]string, error)
	GetTenantPage(int, int) ([]string, int, bool, error)
}

//GetLatestSnapshotsByTenant returns slice of JSON blobs for the latest snapshots of all systems owned by a tenant
func (db *DatabaseSession) GetLatestSnapshotsByTenant(tenant string) ([]string, error) {
	return db.RunQuery("SELECT snapshot FROM latest_snapshots_by_tenant WHERE tenant = ?", tenant)
}

//GetSystemsOfTenant returns a list of serial numbers a given tenant has access to
func (db *DatabaseSession) GetSystemsOfTenant(tenant string) ([]string, error) {
	return db.RunQuery("SELECT serial_number FROM defaultks.latest_snapshots_by_tenant WHERE tenant = ?", tenant)

}

//GetSnapshotByTenantSerialNumberAndDate gets the JSON blob of a system at a specified timestamp
func (db *DatabaseSession) GetSnapshotByTenantSerialNumberAndDate(tenant, serialNumberString, time string) (string, error) {
	stamp, err := StringToTimestamp(time)
	if err != nil {
		return "", err
	}

	serialNumber, err := strconv.Atoi(serialNumberString)
	if err != nil {
		return "", err
	}
	snapshots, err := db.RunQuery("SELECT snapshot FROM snapshots_by_serial_number WHERE tenant = ? AND serial_number = ? AND time = ?", tenant, serialNumber, stamp)
	if err != nil {
		return "", err
	}
	return snapshots[0], nil
}

//TimestampFormat is the format to use in formatting and parsing timestamps
const TimestampFormat string = time.RFC1123

//TimestampsToStrings converts a time slice to string slice for convenience
func TimestampsToStrings(times []time.Time) []string {
	timestamps := make([]string, len(times))
	for i, stamp := range times {
		timestamps[i] = stamp.Format(TimestampFormat)
	}
	return timestamps
}

//StringToTimestamp parses a timestamp to a time object for use in queries
func StringToTimestamp(stamp string) (time.Time, error) {
	return time.Parse(TimestampFormat, stamp)
}

//GetValidTimestampsOfSystem returns a slice of strings that represent valid timestamps to index by for a system
func (db *DatabaseSession) GetValidTimestampsOfSystem(tenant, serialNumberString string) ([]time.Time, error) {
	serialNumber, err := strconv.Atoi(serialNumberString)
	if err != nil {
		return nil, err
	}
	// println(tenant, " ", serialNumber)
	iter := db.Session.Query("SELECT time FROM snapshots_by_serial_number WHERE tenant = ? AND serial_number = ?", tenant, serialNumber).Iter()
	stamps := make([]time.Time, 0, iter.NumRows())
	var stamp time.Time
	for iter.Scan(&stamp) {
		stamps = append(stamps, stamp)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return stamps, nil
}

//RunQuery is a helper function to easily get slice of strings that is returned from query, with error handling
func (db *DatabaseSession) RunQuery(query string, args ...interface{}) ([]string, error) {
	iter := db.Session.Query(query, args...).Iter()
	items := make([]string, 0, iter.NumRows())
	var item string
	for iter.Scan(&item) {
		items = append(items, item)
	}
	//	This checks for errors upon closing the iterator
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return items, nil
}

//GetValidTenants is a query that gets an array of valid tenants
func (db *DatabaseSession) GetValidTenants() ([]string, error) {
	var tenants []string
	iter := db.Session.Query("SELECT tenant FROM latest_snapshots_by_tenant").Iter()
	var tenant string
	for iter.Scan(&tenant) {
		tenants = append(tenants, tenant)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return tenants, nil
}

// GetTenantPage gets the next page of the GetValidTenants query given pageSize
func (db *DatabaseSession) GetTenantPage(pageSize int, page int) ([]string, int, bool, error) {
	var tenants []string
	var lastPageState []byte

	var iter *gocql.Iter

	pageReturned := 0
	isLastPage := false
	db.Session.SetPageSize(pageSize)

	for i := 0; i < page; i++ {
		iter = db.Session.Query("SELECT tenant FROM latest_snapshots_by_tenant ALLOW FILTERING").PageState(lastPageState).Iter()
		lastPageState = iter.PageState()

		pageReturned = i + 1
		if len(lastPageState) == 0 && i+1 < page {

			// Is it necessary to throw a DB error or should we just return the last page?
			if err := iter.Close(); err != nil {
				return nil, 0, false, err
			}

			return nil, 0, false, gocql.Error{
				Message: fmt.Sprintf("Tried to get a page past page %d.\n", i+1),
			}

		} else if len(lastPageState) == 0 {
			isLastPage = true
		}
	}

	var tenant string

	for iter.Scan(&tenant) {
		tenants = append(tenants, tenant)
	}

	if err := iter.Close(); err != nil {
		return nil, 0, false, err
	}

	return tenants, pageReturned, isLastPage, nil
}
