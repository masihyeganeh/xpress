# Xpress Decompressor

Decompresses Xpress contents

---

Example for a HTTP client:

```go
import (
    "github.com/masihyeganeh/xpress"
)

if response.Header.Get("content-encoding") == "xpress" {
	decompressed, err := xpress.Decompress(body)
}
```