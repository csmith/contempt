package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v2"
)

// DownloadYaml requests the given url and then attempts to unmarshal the body as YAML into the provided struct.
func DownloadYaml(url string, i interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	return yaml.NewDecoder(r.Body).Decode(i)
}

// DownloadJson requests the given url and then attempts to unmarshal the body as JSON into the provided struct.
func DownloadJson(url string, i interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(i)
}

// DownloadHash downloads the given URL and parses the first hash out of it, assuming it's formatted in line with the
// output of sha256sum. Hashes are assumed to be hexadecimal and an error will be returned if this is not the case.
func DownloadHash(url string) (string, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	hash := strings.ToLower(strings.SplitN(string(b), " ", 2)[0])
	for i := range hash {
		if (hash[i] < 'a' || hash[i] > 'f') && (hash[i] < '0' || hash[i] > '9') {
			return "", fmt.Errorf("invalid has found at address: %s", hash)
		}
	}
	return hash, nil
}

// FindInHtml downloads the HTML page at the given URL and runs the specified CSS selector over it to find nodes.
// The textual content of those nodes is returned.
func FindInHtml(url string, selector string) ([]string, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var results []string
	doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
		results = append(results, selection.Text())
	})
	return results, nil
}

func RegexURLContent(url string, regex string) (string, error) {
	re := regexp.MustCompile(regex)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	result := re.FindSubmatch(body)
	if len(result) == 0 {
		return "", errors.New("no match found")
	}
	return string(result[1]), nil
}
