package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/dao"

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

// EmailCodes methods
func (a *AIPlanMemAPI) SaveEmailCode(userID uuid.UUID, newEmail string) (string, error) {
	if a.isModule {
		return a.ds.EmailCodes.GenCode(userID, newEmail)
	}
	h := http.Header{}
	h.Set("email", newEmail)

	h, err := a.postRequestWithResponseHeader("/emailCodes/"+userID.String(), h)
	return h.Get("code"), err
}

func (a *AIPlanMemAPI) VerifyEmailCode(userID uuid.UUID, email, code string) (bool, error) {
	if a.isModule {
		return a.ds.EmailCodes.VerifyCode(userID, email, code)
	}

	data := dao.EmailCodeData{
		NewEmail:  email,
		Code:      code,
		CreatedAt: time.Now(),
	}
	jsonData, _ := json.Marshal(data)

	resp, err := a.postRequestWithResponse("/emailCodes/"+userID.String()+"/verify", jsonData)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.Header.Get("verify") == "true", nil
}

//-------------------

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

func (a *AIPlanMemAPI) postRequestWithResponseHeader(path string, header http.Header) (http.Header, error) {
	req, err := http.NewRequest("POST", a.addr.ResolveReference(&url.URL{Path: path}).String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil

}

func (a *AIPlanMemAPI) postRequestWithResponse(path string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	return http.Post(a.addr.ResolveReference(&url.URL{Path: path}).String(), "application/json", reader)
}
