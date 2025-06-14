package repositories

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

func fetch(url string) ([]byte, error) {
	buf := new(bytes.Buffer)

	if url != "" {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Bad response status: %s", resp.Status)
		}
		defer resp.Body.Close()

		io.Copy(buf, resp.Body)
	} else {
		f, err := os.Open("example.xls")
		if err != nil {
			return nil, err
		}
		defer f.Close()

		io.Copy(buf, f)
	}

	return buf.Bytes(), nil
}
