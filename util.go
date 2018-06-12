package relay

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func entryToDestination(entry *logical.StorageEntry) (*Destination, error) {
	var d Destination

	if err := json.Unmarshal(entry.Value, &d); err != nil {
		return nil, errwrap.Wrapf("failed to unmarshal destination: {{err}}", err)
	}

	return &d, nil
}

func getFieldValue(fieldName string, data *framework.FieldData) (interface{}, error) {
	value, ok, err := data.GetOkErr(fieldName)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf("could not parse %s: {{err}}", fieldName), err)
	}
	if ok {
		return value, nil
	} else {
		return data.GetDefaultOrZero(fieldName), nil
	}
}

// StrListContains looks for a string in a list of strings.
func StrListContains(haystack []string, needle string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
