package runtime

import (
"bytes"
"encoding/json"
"io"
"net/http"
"path/filepath"
"testing"
"time"
)

func itoa(i int) string {
if i == 0 {
return "0"
}
neg := i < 0
if neg {
i = -i
}
var b [20]byte
pos := len(b)
for i > 0 {
pos--
b[pos] = byte(0x30 + i%10)
i /= 10
}
if neg {
pos--
b[pos] = 0x2d
}
return string(b[pos:])
}

func newTestServer(t *testing.T) (*Server, string) {
t.Helper()
dbPath := filepath.Join(t.TempDir(), "test.db")
srv := NewServer(dbPath)
port := srv.Start()
if port <= 0 {
t.Fatal("server did not start")
}
t.Cleanup(func() { srv.Stop() })
return srv, "http://127.0.0.1:" + itoa(port)
}

func TestServerStartBindsPort(t *testing.T) {
srv, _ := newTestServer(t)
if srv.Port() <= 0 {
t.Fatal("expected bound port")
}
}

func TestServerTokenIsGenerated(t *testing.T) {
srv, _ := newTestServer(t)
tok := srv.Token()
if len(tok) != 64 {
t.Fatalf("expected 64-char token, got %d", len(tok))
}
}

func TestHealthEndpoint(t *testing.T) {
_, base := newTestServer(t)
time.Sleep(100 * time.Millisecond)
resp, err := http.Get(base + "/health")
if err != nil {
t.Fatal(err)
}
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
if string(body) != "ok" {
t.Fatalf("expected ok, got %q", string(body))
}
}

func TestAPIHealthReportsDBStatus(t *testing.T) {
_, base := newTestServer(t)
time.Sleep(100 * time.Millisecond)
resp, err := http.Get(base + "/api/health")
if err != nil {
t.Fatal(err)
}
defer resp.Body.Close()
var m map[string]string
json.NewDecoder(resp.Body).Decode(&m)
if m["db"] != "ok" {
t.Fatalf("expected db=ok, got %q", m["db"])
}
}

func TestUnauthorizedRequestRejected(t *testing.T) {
_, base := newTestServer(t)
time.Sleep(100 * time.Millisecond)
resp, err := http.Get(base + "/api/notes")
if err != nil {
t.Fatal(err)
}
defer resp.Body.Close()
if resp.StatusCode != 401 {
t.Fatalf("expected 401, got %d", resp.StatusCode)
}
}

func TestCreateAndListNote(t *testing.T) {
srv, base := newTestServer(t)
time.Sleep(100 * time.Millisecond)
body := []byte("{\"title\":\"t1\",\"body\":\"b1\"}")
req, _ := http.NewRequest("POST", base+"/api/notes", bytes.NewReader(body))
req.Header.Set("Authorization", "Bearer "+srv.Token())
resp, err := http.DefaultClient.Do(req)
if err != nil {
t.Fatal(err)
}
resp.Body.Close()
if resp.StatusCode != 201 {
t.Fatalf("expected 201, got %d", resp.StatusCode)
}
req, _ = http.NewRequest("GET", base+"/api/notes", nil)
req.Header.Set("Authorization", "Bearer "+srv.Token())
resp, err = http.DefaultClient.Do(req)
if err != nil {
t.Fatal(err)
}
got, _ := io.ReadAll(resp.Body)
resp.Body.Close()
if !bytes.Contains(got, []byte("t1")) {
t.Fatalf("expected note in list, got %q", string(got))
}
}
