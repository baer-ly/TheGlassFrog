package browser

import (
	"bufio"
	"os"
	"strings"
	"sync/atomic"
)

type ProxyRotator struct {
	proxies []string
	count   uint64
}

func NewProxyRotator(filepath string) (*ProxyRotator, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return &ProxyRotator{proxies: []string{}}, nil
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			proxies = append(proxies, line)
		}
	}
	return &ProxyRotator{proxies: proxies}, nil
}
func (proxy *ProxyRotator) Rotate() string {
	if len(proxy.proxies) == 0 {
		return ""
	}
	idx := atomic.AddUint64(&proxy.count, 1) % uint64(len(proxy.proxies))
	return proxy.proxies[idx]
}
