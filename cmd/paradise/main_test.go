package main

import (
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestDevDefaults(t *testing.T) {
	c, err := loadConfig(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "8080" || c.Production || c.TrustProxy || c.SiteURL != "http://localhost:8080" {
		t.Fatalf("%+v", c)
	}
}

func TestProductionRequiresSettings(t *testing.T) {
	good := map[string]string{
		"APP_ENV": "production", "DATABASE_URL": "postgres://x", "VISITOR_SECRET": strings.Repeat("s", 32),
		"SITE_URL": "https://gamers-paradise.up.railway.app/",
	}
	c, err := loadConfig(env(good))
	if err != nil {
		t.Fatal(err)
	}
	if !c.TrustProxy || c.SiteURL != "https://gamers-paradise.up.railway.app" {
		t.Fatalf("%+v", c)
	}
	for _, drop := range []string{"DATABASE_URL", "VISITOR_SECRET", "SITE_URL"} {
		m := map[string]string{}
		for k, v := range good {
			if k != drop {
				m[k] = v
			}
		}
		if _, err := loadConfig(env(m)); err == nil {
			t.Errorf("production without %s accepted", drop)
		}
	}
	short := map[string]string{}
	for k, v := range good {
		short[k] = v
	}
	short["VISITOR_SECRET"] = "short"
	if _, err := loadConfig(env(short)); err == nil {
		t.Error("short secret accepted")
	}
}

func TestTrustProxyOverride(t *testing.T) {
	c, _ := loadConfig(env(map[string]string{"TRUST_PROXY": "true"}))
	if !c.TrustProxy {
		t.Fatal("override ignored")
	}
}
