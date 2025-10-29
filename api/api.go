package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/aisa-it/aiplan-mem/internal/dao"
	"io"
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

// EmailCodes methods
func (a *AIPlanMemAPI) SaveEmailCode(userID uuid.UUID, newEmail string) (*dao.EmailCodeData, error) {
	if a.isModule {
		return a.ds.EmailCodes.GenCode(userID, newEmail)
	}

	resp, err := a.postRequestWithResponse("/emailCodes/"+userID.String()+"?email="+url.QueryEscape(newEmail), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var codeData dao.EmailCodeData
	if err := json.NewDecoder(resp.Body).Decode(&codeData); err != nil {
		return nil, err
	}

	return &codeData, nil
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

	resp, err := a.postRequestWithResponse("/emailCodes/"+userID.String()+"/verify/", jsonData)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
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

func (a *AIPlanMemAPI) getRequestWithBody(path string) (*http.Response, error) {
	return http.Get(a.addr.ResolveReference(&url.URL{Path: path}).String())
}

func (a *AIPlanMemAPI) postRequest(path string) error {
	resp, err := http.Post(a.addr.ResolveReference(&url.URL{Path: path}).String(), "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (a *AIPlanMemAPI) postRequestWithBody(path string, body []byte) error {
	resp, err := http.Post(a.addr.ResolveReference(&url.URL{Path: path}).String(), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (a *AIPlanMemAPI) postRequestWithResponse(path string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	return http.Post(a.addr.ResolveReference(&url.URL{Path: path}).String(), "application/json", reader)
}
