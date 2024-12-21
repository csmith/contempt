package sources

import (
	"errors"
	"io"
	"net/http"
	"regexp"
)

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
