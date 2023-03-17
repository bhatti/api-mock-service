package types

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"time"
)

// Version creates a formatted struct for output
type Version struct {
	Version string    `json:"version,omitempty"`
	Commit  string    `json:"commit,omitempty"`
	Date    string    `json:"date,omitempty"`
	Started time.Time `json:"started,omitempty"`
}

// NewVersion will create a pointer to a new version object
func NewVersion(version string, commit string, date string) *Version {
	return &Version{
		Version: version,
		Commit:  commit,
		Date:    date,
		Started: time.Now(),
	}
}

func (v *Version) String() string {
	if v.Version == "" || v.Version == v.Commit {
		return v.Commit + "_" + v.Date
	}
	return v.Version + "_" + v.Commit + "_" + v.Date
}

// Output will add the versioning code
func (v *Version) Output(shortened bool) string {
	var response string

	if shortened {
		response = v.ToShortened()
	} else {
		response = v.ToJSON()
	}
	return fmt.Sprintf("%s", response)
}

// ToJSON converts the Version into a JSON String
func (v *Version) ToJSON() string {
	bytes, _ := json.Marshal(v)
	return string(bytes) + "\n"
}

// ToShortened converts the Version into a JSON String
func (v *Version) ToShortened() string {
	bytes, _ := yaml.Marshal(v)
	return string(bytes) + "\n"
}
