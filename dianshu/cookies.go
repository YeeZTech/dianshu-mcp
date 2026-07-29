// Package dianshu - see README for details.
//
// Author: zhyyao
package dianshu

import (
	"encoding/json"
	"os"
)

// LoadCookies loads cookies from a JSON file.
func LoadCookies(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}, nil // return empty on file not found
	}
	var cookies map[string]string
	if err := json.Unmarshal(data, &cookies); err != nil {
		return map[string]string{}, nil
	}
	return cookies, nil
}

// SaveCookies saves cookies to a JSON file.
func SaveCookies(path string, cookies map[string]string) error {
	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DeleteCookiesFile removes the cookies file.
func DeleteCookiesFile(path string) error {
	return os.Remove(path)
}
