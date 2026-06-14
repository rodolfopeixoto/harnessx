package yaml

import "os"

func writeFileImpl(path, body string) error {
	return os.WriteFile(path, []byte(body), 0o644)
}
