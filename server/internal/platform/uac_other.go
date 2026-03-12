//go:build !windows

package platform

func UACEnabled() (*bool, error) {
	return nil, nil
}
