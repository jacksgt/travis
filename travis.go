package travis

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Color is a color that represents the state inside the webhook
type Color int

const (
	// Passed is #39AA56
	Passed Color = 3779158
	// Fail is #DB4545
	Fail Color = 14370117
	// InProgress is #EDDE3F
	InProgress = 15588927
	// Cancel is #9D9D9D
	Cancel = 10329501
)

// Payload for travis
type Payload struct {
	ID                int64       `json:"id,omitempty"`
	Number            string      `json:"number,omitempty"`
	Config            *Config     `json:"config,omitempty"`
	Type              string      `json:"type,omitempty"`
	State             string      `json:"state,omitempty"`
	Status            int         `json:"status,omitempty"`
	Result            int         `json:"result,omitempty"`
	StatusMessage     string      `json:"status_message,omitempty"`
	ResultMessage     string      `json:"result_message,omitempty"`
	StartedAt         time.Time   `json:"started_at,omitempty"`
	FinishedAt        time.Time   `json:"finished_at,omitempty"`
	Duration          int         `json:"duration,omitempty"`
	BuildURL          string      `json:"build_url,omitempty"`
	CommitID          int         `json:"commit_id,omitempty"`
	Commit            string      `json:"commit,omitempty"`
	BaseCommit        string      `json:"base_commit,omitempty"`
	HeadCommit        string      `json:"head_commit,omitempty"`
	Branch            string      `json:"branch,omitempty"`
	Message           string      `json:"message,omitempty"`
	CompareURL        string      `json:"compare_url,omitempty"`
	CommitedAt        time.Time   `json:"commited_at,omitempty"`
	AuthorName        string      `json:"author_name,omitempty"`
	AuthorEmail       string      `json:"author_email,omitempty"`
	CommiterName      string      `json:"commiter_name,omitempty"`
	CommiterEmail     string      `json:"commiter_email,omitempty"`
	PullRequest       int         `json:"pull_request,omitempty"`
	PullRequestNumber int         `json:"pull_request_number,omitempty"`
	PullRequestTitle  string      `json:"pull_request_title,omitempty"`
	Tag               string      `json:"tag,omitempty"`
	Repository        *Repository `json:"repository,omitempty"`
}

// Config field of the payload
type Config struct {
	Sudo     bool   `json:"sudo,omitempty"`
	Dist     string `json:"dist,omitempty"`
	Language string `json:"language,omitempty"`
}

// Repository field of the payload
type Repository struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
	URL       string `json:"url,omitempty"`
}

var travisPubKey *rsa.PublicKey

// GetPayload will parse the payload inside r
func GetPayload(r io.Reader) (*Payload, error) {
	if r == nil {
		return nil, errors.New("can't parse from nil reader")
	}
	p := new(Payload)
	err := json.NewDecoder(r).Decode(p)
	if err != nil {
		return nil, errors.New("cannot decode payload")
	}
	return p, nil
}

// GetPayloadFromRequest will verify the integrity of the request and then
// parse the payload inside the body
func GetPayloadFromRequest(r *http.Request) (*Payload, error) {

	if r.Method != "POST" {
		return nil, fmt.Errorf("wrong request method %q instead of POST", r.Method)
	}

	if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		return nil, fmt.Errorf("wrong Content-Type header, got %s != want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
	}

	key, err := travisPublicKey()
	if err != nil {
		return nil, err
	}

	signature, err := parsePayloadSignature(r)
	if err != nil {
		return nil, err
	}

	payload := payloadDigest(r.FormValue("payload"))

	err = rsa.VerifyPKCS1v15(key, crypto.SHA1, payload, signature)
	if err != nil {
		return nil, errors.New("unauthorized payload")
	}

	p := new(Payload)
	err = json.Unmarshal([]byte(r.FormValue("payload")), p)
	if err != nil {
		return nil, errors.New("cannot decode payload")
	}

	return p, nil
}

type configKey struct {
	Config struct {
		Host        string `json:"host"`
		ShortenHost string `json:"shorten_host"`
		Assets      struct {
			Host string `json:"host"`
		} `json:"assets"`
		Pusher struct {
			Key string `json:"key"`
		} `json:"pusher"`
		Github struct {
			APIURL string   `json:"api_url"`
			Scopes []string `json:"scopes"`
		} `json:"github"`
		Notifications struct {
			Webhook struct {
				PublicKey string `json:"public_key"`
			} `json:"webhook"`
		} `json:"notifications"`
	} `json:"config"`
}

func travisPublicKey() (*rsa.PublicKey, error) {
	/* check if TravisCI's public key is already stored locally */
	if travisPubKey != nil {
		return travisPubKey, nil
	}

	response, err := http.Get("https://api.travis-ci.org/config")

	if err != nil {
		return nil, errors.New("cannot fetch travis public key")
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	var t configKey
	err = decoder.Decode(&t)
	if err != nil {
		return nil, errors.New("cannot decode travis public key")
	}

	key, err := parsePublicKey(t.Config.Notifications.Webhook.PublicKey)
	if err != nil {
		return nil, err
	}

	/* store public key locally */
	travisPubKey = key

	return travisPubKey, nil
}

func parsePublicKey(key string) (*rsa.PublicKey, error) {

	// https://golang.org/pkg/encoding/pem/#Block
	block, _ := pem.Decode([]byte(key))

	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("invalid public key")
	}

	return publicKey.(*rsa.PublicKey), nil

}

func parsePayloadSignature(r *http.Request) ([]byte, error) {
	signature := r.Header.Get("Signature")
	if signature == "" {
		return nil, errors.New("missing Signature header")
	}
	b64, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return nil, errors.New("cannot decode signature")
	}
	return b64, nil
}

func payloadDigest(payload string) []byte {
	hash := sha1.New()
	hash.Write([]byte(payload))
	return hash.Sum(nil)
}

// Pending returns true if a build has been requested
func (p *Payload) Pending() bool {
	return p.StatusMessage == "Pending" || p.ResultMessage == "Pending"
}

// Passed returns true if the build completed successfully
func (p *Payload) Passed() bool {
	return p.StatusMessage == "Passed" || p.ResultMessage == "Passed"
}

// Fixed returns true if the build completed successfully after a previously failed build
func (p *Payload) Fixed() bool {
	return p.StatusMessage == "Fixed" || p.ResultMessage == "Fixed"
}

// Broken returns true if the build completed in failure after a previously successful build
func (p *Payload) Broken() bool {
	return p.StatusMessage == "Broken" || p.ResultMessage == "Broken"
}

// Failed returns true if the build is the first build for a new branch and has failed
func (p *Payload) Failed() bool {
	return p.StatusMessage == "Failed" || p.ResultMessage == "Failed"
}

// StillFailing returns true if the build completed in failure after a previously failed build
func (p *Payload) StillFailing() bool {
	return p.StatusMessage == "Still Failing" || p.ResultMessage == "Still Failing"
}

// Canceled returns true if the build was canceled
func (p *Payload) Canceled() bool {
	return p.StatusMessage == "Canceled" || p.ResultMessage == "Canceled"
}

// Errored returns true if the build has errored
func (p *Payload) Errored() bool {
	return p.StatusMessage == "Errored" || p.ResultMessage == "Errored"
}

// IsPullRequest returns true if the event was caused by a pull request
func (p *Payload) IsPullRequest() bool {
	return p.Type == "pull_request"
}

// IsPush returns true if the event was caused by a push
func (p *Payload) IsPush() bool {
	return p.Type == "push"
}

// IsCron returns true if the event was caused by a cron
func (p *Payload) IsCron() bool {
	return p.Type == "cron"
}

// IsAPI returns true if the event was caused by an api
func (p *Payload) IsAPI() bool {
	return p.Type == "api"
}
