package get_env

import (
	"fmt"
	"os"
	"strings"
)

// GetServiceAddr возвращает адрес сервиса в зависимости от среды исполнения.
// Если RUNNING_IN_CONTAINER=true, то используется значение переменной окружения.
// Иначе — fallback (обычно localhost:порт).
func GetServiceAddr(serviceVar string, defaultPort int) string {
	if strings.ToLower(os.Getenv("RUNNING_IN_CONTAINER")) == "true" {
		if val := os.Getenv(serviceVar); val != "" {
			return val
		}
	}
	return fmt.Sprintf("127.0.0.1:%d", defaultPort)
}
