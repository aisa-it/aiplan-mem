package api

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/aisa-it/aiplan-mem/internal/db"
	"github.com/gofrs/uuid/v5"
)

type AIPlanMemAPI struct {
	ds       *db.DataStore
	isModule bool
	addr     *url.URL
}

func NewClient(isModule bool, pathOrAddr string) (*AIPlanMemAPI, error) {
	var err error
	a := &AIPlanMemAPI{isModule: isModule}
	if isModule {
		cfg := &config.Config{Path: pathOrAddr}
		a.ds, err = db.OpenDB(cfg)
		if err != nil {
			return nil, err
		}
	} else {
		a.addr, err = url.Parse(pathOrAddr)
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

func (a *AIPlanMemAPI) Close() error {
	return a.ds.Close()
}

func (a *AIPlanMemAPI) BlacklistToken(signature []byte) error {
	if a.isModule {
		return a.ds.Sessions.BlacklistToken(signature)
	}
	return a.postRequest("/blacklist/" + base64.StdEncoding.EncodeToString(signature))
}
func (a *AIPlanMemAPI) IsTokenBlacklisted(signature []byte) (bool, error) {
	if a.isModule {
		return a.ds.Sessions.IsTokenBlacklisted(signature)
	}
	h, err := a.getRequest("/blacklist/" + base64.StdEncoding.EncodeToString(signature))
	return h.Get("blacklisted") == "true", err
}
func (a *AIPlanMemAPI) SaveUserLastSeenTime(userId uuid.UUID) error {
	if a.isModule {
		return a.ds.Sessions.SaveUserLastSeenTime(userId)
	}
	return a.postRequest("/lastSeen/" + userId.String())
}
func (a *AIPlanMemAPI) GetUserLastSeenTime(userId uuid.UUID) (time.Time, error) {
	if a.isModule {
		return a.ds.Sessions.GetUserLastSeenTime(userId)
	}
	h, err := a.getRequest("/lastSeen/" + userId.String())

	t, err := strconv.Atoi(h.Get("LastSeen"))
	return time.Unix(int64(t), 0), err
}

func (a *AIPlanMemAPI) getRequest(path string) (http.Header, error) {
	resp, err := http.Get(a.addr.ResolveReference(&url.URL{Path: path}).String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func (a *AIPlanMemAPI) postRequest(path string) error {
	resp, err := http.Post(a.addr.ResolveReference(&url.URL{Path: path}).String(), "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
