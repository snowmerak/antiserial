package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/snowmerak/antiserial/test/testgen_v2"
)

func goldenPayloadV2Path(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "golden", "payload_v2.bin")
}

func TestGoldenPayloadV2MatchesGoMarshal(t *testing.T) {
	p := testgen_v2.Payload{
		Id:     1234567890,
		Uuid:   "abc",
		Active: true,
		Tags:   []string{"go", "rust"},
	}
	got, err := p.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}

	golden, err := os.ReadFile(goldenPayloadV2Path(t))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, golden) {
		t.Fatalf("marshal mismatch:\n got %x\nwant %x", got, golden)
	}
}

func TestGoldenPayloadV2GoDecode(t *testing.T) {
	golden, err := os.ReadFile(goldenPayloadV2Path(t))
	if err != nil {
		t.Fatal(err)
	}
	var p testgen_v2.Payload
	n, err := p.Unmarshal(golden)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(golden) {
		t.Fatalf("consumed %d of %d", n, len(golden))
	}
	if p.Id != 1234567890 || p.Uuid != "abc" || !p.Active {
		t.Fatalf("fields: %+v", p)
	}
}

func TestGoldenPayloadV2PythonDecode(t *testing.T) {
	goldenPath := goldenPayloadV2Path(t)
	script := filepath.Join(filepath.Dir(goldenPath), "..", "golden", "verify_v2.py")
	cmd := exec.Command("python", script, goldenPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python verify: %v\n%s", err, out)
	}
}

func TestGoldenPayloadV2RustDecode(t *testing.T) {
	goldenPath := goldenPayloadV2Path(t)
	rustDir := filepath.Join(filepath.Dir(goldenPath), "..", "test_rust")
	cmd := exec.Command("cargo", "run", "--quiet", "--", goldenPath)
	cmd.Dir = rustDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("rust verify: %v\n%s", err, out)
	}
}

func TestGoldenPayloadV2TypeScriptDecode(t *testing.T) {
	goldenPath := goldenPayloadV2Path(t)
	script := filepath.Join(filepath.Dir(goldenPath), "..", "test_ts", "verify_v2.ts")
	cmd := exec.Command("deno", "run", "--allow-read=" + filepath.Dir(goldenPath), script, goldenPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("deno verify: %v\n%s", err, out)
	}
}