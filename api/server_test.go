package api

import (
	"testing"
	"strings"
	"net/http"
	"net/http/httptest"
	"time"
)

type mockSession struct{}


func (db *mockSession) GetSystemsOfTenant(tenant string) ([]string, error) {
	return []string{"9996788"}, nil
}

func (db *mockSession) GetLatestSnapshotsByTenant(tenant string) ([]string, error) {

	snapshots := []string{
		"{ SerialNumberInserv: \"1231234\", tenant: { authorized:[\"%TENANT%\"]}",
		"{ SerialNumberInserv: \"7162634\", tenant: { authorized:[\"%TENANT%\"]}",
		"{ SerialNumberInserv: \"1111111\", tenant: { authorized:[\"%TENANT%\"]}",
	}
	for i, s := range snapshots {
		snapshots[i] = strings.Replace(s, "%TENANT%", tenant, 1)
	}

	return snapshots, nil
}

func (db *mockSession) GetSnapshotByTenantSerialNumberAndDate(tenant, time, serialNumber string) (string, error) {
	snapshot := "{ SerialNumberInserv: \"%SYSID%\", tenant: { authorized:[\"%TENANT%\"]}"
	snapshot = strings.Replace(snapshot, "%TENANT%", tenant, 1)
	snapshot = strings.Replace(snapshot, "%SYSID%", serialNumber, 1)
	return snapshot, nil
}

func (db *mockSession) GetValidTimestampsOfSystem(tenant, serialNumber string) ([]time.Time, error) {
	return []time.Time{time.Now()}, nil
}

func TestGetSerialNumbersOfTenant(t *testing.T) {
	const expectedObtainedString = "\n...expected = %#v\n...obtained = %#v"

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tenants/hpe/systems", nil)
	env := Server{
		DBSession: &mockSession{},
	}
	env.getTenantSystems().ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Errorf(expectedObtainedString, http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf(expectedObtainedString, "application/json", rec.Header().Get("Content-Type"))
	}

	expected := "[\"9996788\"]"
	if expected != rec.Body.String() {
		t.Errorf(expectedObtainedString, expected, rec.Body.String())
	}
}


func TestHandleGetLatestSnapshotsByTenantWithTextTenant(t *testing.T) {
	const expectedObtainedString = "\n...expected = %#v\n...obtained = %#v"

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tenants/hpe/snapshots", nil)

	env := Server{
		DBSession: &mockSession{},
	}
	env.getLatestSnapshotsByTenant().ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Errorf(expectedObtainedString, http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf(expectedObtainedString, "application/json", rec.Header().Get("Content-Type"))
	}

	expected :=
		"[{ SerialNumberInserv: \"1231234\", tenant: { authorized:[\"hpe\"]}," +
			"{ SerialNumberInserv: \"7162634\", tenant: { authorized:[\"hpe\"]}," +
			"{ SerialNumberInserv: \"1111111\", tenant: { authorized:[\"hpe\"]}]"
	if expected != rec.Body.String() {
		t.Errorf(expectedObtainedString, expected, rec.Body.String())
	}
}

func TestHandleGetLatestSnapshotsByTenantWithTenantID(t *testing.T) {
	const expectedObtainedString = "\n...expected = %#v\n...obtained = %#v"

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/tenants/264593856/snapshots", nil)

	env := Server{
		DBSession: &mockSession{},
	}
	env.getLatestSnapshotsByTenant().ServeHTTP(rec, req)

	if http.StatusOK != rec.Code {
		t.Errorf(expectedObtainedString, http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf(expectedObtainedString, "application/json", rec.Header().Get("Content-Type"))
	}

	expected :=
		"[{ SerialNumberInserv: \"1231234\", tenant: { authorized:[\"264593856\"]}," +
			"{ SerialNumberInserv: \"7162634\", tenant: { authorized:[\"264593856\"]}," +
			"{ SerialNumberInserv: \"1111111\", tenant: { authorized:[\"264593856\"]}]"
	if expected != rec.Body.String() {
		t.Errorf(expectedObtainedString, expected, rec.Body.String())
	}
}