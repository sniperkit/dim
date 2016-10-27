package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
	"github.com/howeyc/gopass"
)

// ListContains checks a list of string contains a given string
func ListContains(list []string, search string) bool {
	for _, r := range list {
		if r == search {
			return true
		}
	}
	return false
}

// MapContainsNone checks a map of string does not contain any given string as key
/*func MapContainsNone(all map[string]string, forbidden []string) bool {
	if all == nil || len(all) == 0 || forbidden == nil || len(forbidden) == 0 {
		return true
	}
	for _, f := range forbidden {
		if all[f] != "" {
			return false
		}
	}
	return true
}

// SelectMapKeys retuns key of a map that exist in the keys list and that have non-empty values
func SelectMapKeys(all map[string]string, keys []string) []string {
	if all == nil || len(all) == 0 || keys == nil || len(keys) == 0 {
		return nil
	}

	selected := make([]string, len(keys))

	for _, k := range keys {
		if all[k] != "" {
			selected = append(selected, k)
		}
	}
	return selected

}*/

// MapMatchesAll checks first map contains all the second map elements with the same value
func MapMatchesAll(all, search map[string]string) bool {

	if all == nil || search == nil {
		return false
	}

	if len(all) >= len(search) {
		for k, v := range search {
			if all[k] != v {
				return false
			}
		}
		return true
	}
	return false

}

// Keys returns all the keys of the given map
func Keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// FilterValues removes all element from m where value is s
func FilterValues(m map[string]string, s string) map[string]string {
	filtered := make(map[string]string, len(m))
	for k, v := range m {
		if v != s {
			filtered[k] = v
		}
	}
	return filtered

}

// ReadCredentials ask the user to enter his login and password and stores the value in the given registryAuth
func ReadCredentials(registryAuth *types.AuthConfig) {
	logrus.WithFields(logrus.Fields{"Login": registryAuth.Username, "Password": registryAuth.Password}).Debugln("Reading new credentials")
	if registryAuth.Username != "" {
		fmt.Printf("Username (%s) :", registryAuth.Username)
	} else {
		fmt.Print("Username :")
	}
	var input string
	fmt.Scanln(&input)
	if input != "" {
		registryAuth.Username = input
	} else if registryAuth.Username == "" {
		return
	}
	fmt.Print("Password :")
	pwd, _ := gopass.GetPasswd()
	input = string(pwd)
	if input == "" {
		return
	}
	registryAuth.Password = input
}

// BuildURL appends http:// or https:// to given hostname, according to the insecure parameter
func BuildURL(hostname string, insecure bool) string {
	if hostname == "" {
		return ""
	}
	if strings.HasPrefix(hostname, "http://") || strings.HasPrefix(hostname, "https://") {
		return hostname
	}

	protocol := "https"
	if insecure {
		protocol = "http"
	}
	return fmt.Sprintf("%s://%s", protocol, hostname)

}

// FlatMap returns a string representation of a map
func FlatMap(m map[string]string) string {
	if m == nil || len(m) == 0 {
		return ""
	}
	entries := make([]string, 0, len(m))
	for k, v := range m {
		entries = append(entries, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(entries)
	return strings.Join(entries, ", ")
}
