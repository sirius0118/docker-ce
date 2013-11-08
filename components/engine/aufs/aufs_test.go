package aufs

import (
	"os"
	"path"
	"testing"
)

var (
	tmp = path.Join(os.TempDir(), "aufs-tests")
)

func newDriver(t *testing.T) *AufsDriver {
	if err := os.MkdirAll(tmp, 0755); err != nil {
		t.Fatal(err)
	}

	d, err := Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	return d.(*AufsDriver)
}

func TestNewAufsDriver(t *testing.T) {
	if err := os.MkdirAll(tmp, 0755); err != nil {
		t.Fatal(err)
	}

	d, err := Init(tmp)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	if d == nil {
		t.Fatalf("Driver should not be nil")
	}
}

func TestAufsString(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if d.String() != "aufs" {
		t.Fatalf("Expected aufs got %s", d.String())
	}
}

func TestCreateDirStructure(t *testing.T) {
	newDriver(t)
	defer os.RemoveAll(tmp)

	paths := []string{
		"mnt",
		"layers",
		"diff",
	}

	for _, p := range paths {
		if _, err := os.Stat(path.Join(tmp, "aufs", p)); err != nil {
			t.Fatal(err)
		}
	}
}

// We should be able to create two drivers with the same dir structure
func TestNewDriverFromExistingDir(t *testing.T) {
	if err := os.MkdirAll(tmp, 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := Init(tmp); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(tmp); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(tmp)
}

func TestCreateNewDir(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}
}

func TestCreateNewDirStructure(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		"mnt",
		"diff",
		"layers",
	}

	for _, p := range paths {
		if _, err := os.Stat(path.Join(tmp, "aufs", p, "1")); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRemoveImage(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	if err := d.Remove("1"); err != nil {
		t.Fatal(err)
	}

	paths := []string{
		"mnt",
		"diff",
		"layers",
	}

	for _, p := range paths {
		if _, err := os.Stat(path.Join(tmp, "aufs", p, "1")); err == nil {
			t.Fatalf("Error should not be nil because dirs with id 1 should be delted: %s", p)
		}
	}
}

func TestGetWithoutParent(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	diffPath, err := d.Get("1")
	if err != nil {
		t.Fatal(err)
	}
	expected := path.Join(tmp, "aufs", "diff", "1")
	if diffPath != expected {
		t.Fatalf("Expected path %s got %s", expected, diffPath)
	}
}

func TestCleanupWithNoDirs(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupWithDir(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	if err := d.Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestMountedFalseResponse(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	response, err := d.mounted("1")
	if err != nil {
		t.Fatal(err)
	}

	if response != false {
		t.Fatalf("Response if dir id 1 is mounted should be false")
	}
}

func TestMountedTrueReponse(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)
	defer d.Cleanup()

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}
	if err := d.Create("2", "1"); err != nil {
		t.Fatal(err)
	}

	_, err := d.Get("2")
	if err != nil {
		t.Fatal(err)
	}

	response, err := d.mounted("2")
	if err != nil {
		t.Fatal(err)
	}

	if response != true {
		t.Fatalf("Response if dir id 2 is mounted should be true")
	}
}

func TestMountWithParent(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}
	if err := d.Create("2", "1"); err != nil {
		t.Fatal(err)
	}

	mntPath, err := d.Get("2")
	if err != nil {
		t.Fatal(err)
	}
	if mntPath == "" {
		t.Fatal("mntPath should not be empty string")
	}

	expected := path.Join(tmp, "aufs", "mnt", "2")
	if mntPath != expected {
		t.Fatalf("Expected %s got %s", expected, mntPath)
	}

	if err := d.Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateWithInvalidParent(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", "docker"); err == nil {
		t.Fatalf("Error should not be nil with parent does not exist")
	}
}

func TestGetDiff(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	diffPath, err := d.Get("1")
	if err != nil {
		t.Fatal(err)
	}

	// Add a file to the diff path with a fixed size
	size := int64(1024)

	f, err := os.Create(path.Join(diffPath, "test_file"))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
	f.Close()

	a, err := d.Diff("1")
	if err != nil {
		t.Fatal(err)
	}
	if a == nil {
		t.Fatalf("Archive should not be nil")
	}
}

/* FIXME: How to properly test this?
func TestDiffSize(t *testing.T) {
	d := newDriver(t)
	defer os.RemoveAll(tmp)

	if err := d.Create("1", ""); err != nil {
		t.Fatal(err)
	}

	diffPath, err := d.Get("1")
	if err != nil {
		t.Fatal(err)
	}

	// Add a file to the diff path with a fixed size
	size := int64(1024)

	f, err := os.Create(path.Join(diffPath, "test_file"))
	if err != nil {
		t.Fatal(err)
	}
	f.Truncate(size)
	s, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	size = s.Size()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	diffSize, err := d.DiffSize("1")
	if err != nil {
		t.Fatal(err)
	}
	if diffSize != size {
		t.Fatalf("Expected size to be %d got %d", size, diffSize)
	}
}
*/
