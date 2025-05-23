// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package quick

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestReadVersion(t *testing.T) {
	type myStruct struct {
		Version string
	}
	saveMe := myStruct{"1"}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Save("test.json")
	if err != nil {
		t.Fatal(err)
	}

	version, err := GetVersion("test.json", nil)
	if err != nil {
		t.Fatal(err)
	}
	if version != "1" {
		t.Fatalf("Expected version '1', got '%v'", version)
	}
}

func TestReadVersionErr(t *testing.T) {
	type myStruct struct {
		Version int
	}
	saveMe := myStruct{1}
	_, err := NewConfig(&saveMe, nil)
	if err == nil {
		t.Fatal("Unexpected should fail in initialization for bad input")
	}

	err = os.WriteFile("test.json", []byte("{ \"version\":2,"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetVersion("test.json", nil)
	if err == nil {
		t.Fatal("Unexpected should fail to fetch version")
	}

	err = os.WriteFile("test.json", []byte("{ \"version\":2 }"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetVersion("test.json", nil)
	if err == nil {
		t.Fatal("Unexpected should fail to fetch version")
	}
}

func TestSaveFailOnDir(t *testing.T) {
	defer os.RemoveAll("test-1.json")
	err := os.MkdirAll("test-1.json", 0o644)
	if err != nil {
		t.Fatal(err)
	}
	type myStruct struct {
		Version string
	}
	saveMe := myStruct{"1"}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Save("test-1.json")
	if err == nil {
		t.Fatal("Unexpected should fail to save if test-1.json is a directory")
	}
}

func TestCheckData(t *testing.T) {
	err := CheckData(nil)
	if err == nil {
		t.Fatal("Unexpected should fail")
	}

	type myStructBadNoVersion struct {
		User        string
		Password    string
		Directories []string
	}
	saveMeBadNoVersion := myStructBadNoVersion{"guest", "nopassword", []string{"Work", "Documents", "Music"}}
	err = CheckData(&saveMeBadNoVersion)
	if err == nil {
		t.Fatal("Unexpected should fail if Version is not set")
	}

	type myStructBadVersionInt struct {
		Version  int
		User     string
		Password string
	}
	saveMeBadVersionInt := myStructBadVersionInt{1, "guest", "nopassword"}
	err = CheckData(&saveMeBadVersionInt)
	if err == nil {
		t.Fatal("Unexpected should fail if Version is integer")
	}

	type myStructGood struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}

	saveMeGood := myStructGood{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	err = CheckData(&saveMeGood)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadFile(t *testing.T) {
	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}
	saveMe := myStruct{}
	_, err := LoadConfig("test.json", nil, &saveMe)
	if err == nil {
		t.Fatal(err)
	}

	file, err := os.Create("test.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = LoadConfig("test.json", nil, &saveMe)
	if err == nil {
		t.Fatal("Unexpected should fail to load empty JSON")
	}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Load("test-non-exist.json")
	if err == nil {
		t.Fatal("Unexpected should fail to Load non-existent config")
	}

	err = config.Load("test.json")
	if err == nil {
		t.Fatal("Unexpected should fail to load empty JSON")
	}

	saveMe = myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	config, err = NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Save("test.json")
	if err != nil {
		t.Fatal(err)
	}
	saveMe1 := myStruct{}
	_, err = LoadConfig("test.json", nil, &saveMe1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(saveMe1, saveMe) {
		t.Fatalf("Expected %v, got %v", saveMe1, saveMe)
	}

	saveMe2 := myStruct{}
	err = json.Unmarshal([]byte(config.String()), &saveMe2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(saveMe2, saveMe1) {
		t.Fatalf("Expected %v, got %v", saveMe2, saveMe1)
	}
}

func TestYAMLFormat(t *testing.T) {
	testYAML := "test.yaml"
	defer os.RemoveAll(testYAML)

	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}

	plainYAML := `version: "1"
user: guest
password: nopassword
directories:
    - Work
    - Documents
    - Music
`

	if runtime.GOOS == "windows" {
		plainYAML = strings.ReplaceAll(plainYAML, "\n", "\r\n")
	}

	saveMe := myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}

	// Save format using
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = config.Save(testYAML)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the saved structure in actually an YAML format
	b, err := os.ReadFile(testYAML)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal([]byte(plainYAML), b) {
		t.Fatalf("Expected %v, got %v", plainYAML, string(b))
	}

	// Check if the loaded data is the same as the saved one
	loadMe := myStruct{}
	config, err = NewConfig(&loadMe, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = config.Load(testYAML)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(saveMe, loadMe) {
		t.Fatalf("Expected %v, got %v", saveMe, loadMe)
	}
}

func TestJSONFormat(t *testing.T) {
	testJSON := "test.json"
	defer os.RemoveAll(testJSON)

	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}

	plainJSON := `{
	"Version": "1",
	"User": "guest",
	"Password": "nopassword",
	"Directories": [
		"Work",
		"Documents",
		"Music"
	]
}`

	if runtime.GOOS == "windows" {
		plainJSON = strings.ReplaceAll(plainJSON, "\n", "\r\n")
	}

	saveMe := myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}

	// Save format using
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = config.Save(testJSON)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the saved structure in actually an JSON format
	b, err := os.ReadFile(testJSON)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal([]byte(plainJSON), b) {
		t.Fatalf("Expected %v, got %v", plainJSON, string(b))
	}

	// Check if the loaded data is the same as the saved one
	loadMe := myStruct{}
	config, err = NewConfig(&loadMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Load(testJSON)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(saveMe, loadMe) {
		t.Fatalf("Expected %v, got %v", saveMe, loadMe)
	}
}

func TestSaveLoad(t *testing.T) {
	defer os.RemoveAll("test.json")
	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}
	saveMe := myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Save("test.json")
	if err != nil {
		t.Fatal(err)
	}

	loadMe := myStruct{Version: "1"}
	newConfig, err := NewConfig(&loadMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = newConfig.Load("test.json")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config.Data(), newConfig.Data()) {
		t.Fatalf("Expected %v, got %v", config.Data(), newConfig.Data())
	}
	if !reflect.DeepEqual(config.Data(), &loadMe) {
		t.Fatalf("Expected %v, got %v", config.Data(), &loadMe)
	}

	mismatch := myStruct{"1.1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	if reflect.DeepEqual(config.Data(), &mismatch) {
		t.Fatal("Expected to mismatch but succeeded instead")
	}
}

func TestSaveBackup(t *testing.T) {
	defer os.RemoveAll("test.json")
	defer os.RemoveAll("test.json.old")
	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}
	saveMe := myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Save("test.json")
	if err != nil {
		t.Fatal(err)
	}

	loadMe := myStruct{Version: "1"}
	newConfig, err := NewConfig(&loadMe, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = newConfig.Load("test.json")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config.Data(), newConfig.Data()) {
		t.Fatalf("Expected %v, got %v", config.Data(), newConfig.Data())
	}
	if !reflect.DeepEqual(config.Data(), &loadMe) {
		t.Fatalf("Expected %v, got %v", config.Data(), &loadMe)
	}

	mismatch := myStruct{"1.1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	if reflect.DeepEqual(newConfig.Data(), &mismatch) {
		t.Fatal("Expected to mismatch but succeeded instead")
	}

	config, err = NewConfig(&mismatch, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = config.Save("test.json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDiff(t *testing.T) {
	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}
	saveMe := myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}

	type myNewConfigStruct struct {
		Version string
		// User     string
		Password    string
		Directories []string
	}

	mismatch := myNewConfigStruct{"1", "nopassword", []string{"Work", "documents", "Music"}}
	newConfig, err := NewConfig(&mismatch, nil)
	if err != nil {
		t.Fatal(err)
	}

	fields, err := config.Diff(newConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 {
		t.Fatalf("Expected len 1, got %v", len(fields))
	}

	// Uncomment for debugging
	//	for i, field := range fields {
	//		fmt.Printf("Diff[%d]: %s=%v\n", i, field.Name(), field.Value())
	//	}
}

func TestDeepDiff(t *testing.T) {
	type myStruct struct {
		Version     string
		User        string
		Password    string
		Directories []string
	}
	saveMe := myStruct{"1", "guest", "nopassword", []string{"Work", "Documents", "Music"}}
	config, err := NewConfig(&saveMe, nil)
	if err != nil {
		t.Fatal(err)
	}

	mismatch := myStruct{"1", "Guest", "nopassword", []string{"Work", "documents", "Music"}}
	newConfig, err := NewConfig(&mismatch, nil)
	if err != nil {
		t.Fatal(err)
	}

	fields, err := config.DeepDiff(newConfig)
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("Expected len 2, got %v", len(fields))
	}

	// Uncomment for debugging
	//	for i, field := range fields {
	//		fmt.Printf("DeepDiff[%d]: %s=%v\n", i, field.Name(), field.Value())
	//	}
}
