package config

import (
	"fmt"
	"os"
	"testing"
)

func cleanup() {
	os.Remove("/Users/mrodgers/.questconfig_test")
}

func TestMain(m *testing.M) {
	m.Run()
	cleanup()
}

func TestOpenConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}
	configFile.Close()
}

func TestGetFromEmptyConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	value, err := configFile.GetValue("TEST")
	if err != nil {
		t.Fatal("get error: ", err)
	}
	if value != "" {
		t.Fatalf("expected \"\", got \"%s\"", value)
	}
}

func TestWriteEmptyConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	fullValue, err := configFile.AppendValue("TEST", "my test value")
	if err != nil {
		t.Fatal("write error: ", err)
	}
	if fullValue != "my test value" {
		t.Fatalf("expected return value to equal \"my test value\", got \"%s\"", fullValue)
	}

	wantConfigData := "TEST=my test value"
	configData, _ := configFile.readData()
	if string(configData) != wantConfigData {
		t.Fatalf("expect file data to equal \"%s\", got \"%s\"", wantConfigData, configData)
	}

	configFile.Close()
}

func TestOpenExistingConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	wantConfigData := "TEST=my test value"
	configData, _ := configFile.readData()
	if string(configData) != wantConfigData {
		t.Fatalf("expect file data to equal \"%s\", got \"%s\"", wantConfigData, configData)
	}

	configFile.Close()
}

func TestGetFromExistingConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	wantValue := "my test value"
	value, err := configFile.GetValue("TEST")
	if err != nil {
		t.Fatal("get error: ", err)
	}
	if value != wantValue {
		t.Fatalf("expected \"%s\", got \"%s\"", wantValue, value)
	}
}

func TestWriteExistingConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	wantFullValue := "a new entry"
	fullValue, err := configFile.AppendValue("OTHER", "a new entry")
	if err != nil {
		t.Fatal("write error: ", err)
	}
	if fullValue != "a new entry" {
		t.Fatalf("expected return value to equal \"%s\", got \"%s\"", wantFullValue, fullValue)
	}

	wantConfigData := "TEST=my test value\nOTHER=a new entry"
	configData, _ := configFile.readData()
	if string(configData) != wantConfigData {
		t.Fatalf("expect file data to equal \"%s\", got \"%s\"", wantConfigData, configData)
	}

	configFile.Close()
}

func TestAppendExistingConfigFile(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	tests := []struct {
		k, v                   string
		wantReturned, wantFull string
	}{
		{
			"TEST",
			"an appended value",
			"my test value; an appended value",
			"TEST=my test value; an appended value\nOTHER=a new entry",
		},
		{
			"OTHER",
			"and one more",
			"a new entry; and one more",
			"TEST=my test value; an appended value\nOTHER=a new entry; and one more",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s,%s", tt.k, tt.v), func(t *testing.T) {
			returned, err := configFile.AppendValue(tt.k, tt.v)
			if err != nil {
				t.Fatal("write error: ", err)
			}
			if returned != tt.wantReturned {
				t.Fatalf("expected return value to equal \"%s\", got \"%s\"", tt.wantReturned, returned)
			}

			if full, _ := configFile.readData(); string(full) != tt.wantFull {
				t.Fatalf("expect file data to equal \"%s\", got \"%s\"", tt.wantFull, full)
			}
		})
	}

	configFile.Close()
}

func TestRemoveExistingConfig(t *testing.T) {
	configFile, err := OpenConfigFile(".questconfig_test")
	if err != nil {
		t.Fatal("Could not open config file: ", err)
	}

	tests := []struct {
		k            string
		wantReturned bool
		wantFull     string
	}{
		{"OTHER", true, "TEST=my test value; an appended value"},
		{"TEST", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.k, func(t *testing.T) {
			returned, err := configFile.RemoveValue(tt.k)
			if err != nil {
				t.Fatal("remove error: ", err)
			}
			if returned != tt.wantReturned {
				t.Fatalf("expected return value to equal \"%t\", got \"%t\"", tt.wantReturned, returned)
			}
			if full, _ := configFile.readData(); string(full) != tt.wantFull {
				t.Fatalf("expect file data to equal \"%s\", got \"%s\"", tt.wantFull, full)
			}
		})
	}

	configFile.Close()
}
