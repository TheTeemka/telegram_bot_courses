package courses

import (
	"bytes"
	"io"
	"os"
)

func fetch(url string) ([]byte, error) {
	// resp, err := http.Get(url)
	// if err != nil {
	// 	panic(err)
	// }
	// if resp.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("Bad response status: %s", resp.Status)
	// }
	// defer resp.Body.Close()

	f, err := os.Open("ala.xls")
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, f)
	// io.Copy(buf, resp.Body)

	return buf.Bytes(), nil
}
