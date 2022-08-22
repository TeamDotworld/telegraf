//go:build linux
// +build linux

package wmi_win

func CreateQuery(src interface{}, where string, class ...string) string {
	return ""
}

func Query(query string, dst interface{}, connectServerArgs ...interface{}) error {
	return nil
}
