package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
)

type Store struct {
	Dir string
}

func NewStore(dir string) (*Store, error) {
	dir, err := homedir.Expand(dir)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(dir, string(os.PathSeparator)) {
		dir += string(os.PathSeparator)
	}
	return &Store{Dir: dir}, nil
}

func (s *Store) Location(loc string, redact bool) (string, bool) {
	if filepath.IsAbs(loc) || strings.HasPrefix(loc, "."+string(os.PathSeparator)) {
		return loc, false
	}
	u, err := url.Parse(loc)
	if err != nil || u.Scheme == "" {
		return filepath.Join(s.Dir, loc), false
	}
	if redact && u.RawQuery != "" {
		u.RawQuery = "..."
	}
	return u.String(), true
}

func (s *Store) ReadFile(loc string) ([]byte, error) {
	aLoc, isURL := s.Location(loc, false)
	if isURL {
		u, err := url.Parse(aLoc)
		if err != nil {
			return nil, err
		}
		switch u.Scheme {
		case "https", "http":
			res, err := http.Get(aLoc)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("%s", res.Status)
			}
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			return b, nil
		}
		return nil, fmt.Errorf("Unsupported location to load")
	}
	return ioutil.ReadFile(aLoc)
}

func (s *Store) WriteFile(loc string, b []byte, m os.FileMode) error {
	aLoc, isURL := s.Location(loc, false)
	if isURL {
		u, err := url.Parse(aLoc)
		if err != nil {
			return err
		}
		switch u.Scheme {
		case "https":
			if strings.HasSuffix(u.Host, ".blob.core.windows.net") {
				cli := &http.Client{}
				req, err := http.NewRequest(http.MethodPut, aLoc, bytes.NewBuffer(b))
				if err != nil {
					return err
				}
				req.Header.Set("x-ms-blob-type", "BlockBlob")
				res, err := cli.Do(req)
				if err != nil {
					return err
				}
				defer res.Body.Close()
				if res.StatusCode != http.StatusCreated {
					return fmt.Errorf("%s", res.Status)
				}
				return nil
			}
		}
		return fmt.Errorf("Unsupported location to save")
	}
	if strings.HasPrefix(aLoc, s.Dir) {
		err := os.MkdirAll(filepath.Dir(s.Dir), 0755)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(aLoc, b, m)
}
