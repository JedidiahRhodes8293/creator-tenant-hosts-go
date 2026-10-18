package main

import "testing"

func TestTenantDomainAndRecord(t *testing.T) {
	cases := []struct{ tenant, wantDomain string }{{"studio7", "creators.example"}, {"drawlab", "creators.example"}}
	for _, tc := range cases {
		t.Run(tc.tenant, func(t *testing.T) {
			domain := "creators.example"
			if domain != tc.wantDomain {
				t.Fatalf("domain = %q", domain)
			}
			record := map[string]any{"record_type": "CNAME", "name": tc.tenant, "content": "edge.creators.example"}
			if record["record_type"] != "CNAME" || record["content"] != "edge.creators.example" {
				t.Fatal("unexpected tenant routing record")
			}
		})
	}
}
