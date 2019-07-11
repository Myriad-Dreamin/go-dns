package mdnet

import "fmt"

const (
	MAX_PORT_NUM = 65535
)

func ConvertPortFromInt(port int) (string, error) {
	if port > MAX_PORT_NUM {
		return "", fmt.Errorf("input port exceed max port number")
	}
	return ":" + string(port), nil
}
