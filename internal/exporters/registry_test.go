package exporters

import "testing"

func TestDiscoverFindsCSVManifest(t *testing.T) {
	registry, err := Discover()
	if err != nil {
		t.Fatal(err)
	}
	format, ok := registry.Format("csv")
	if !ok || !format.SupportsColumns || format.Labels["zh-CN"] != "CSV" {
		t.Fatalf("CSV format = %#v, found = %v", format, ok)
	}
}
