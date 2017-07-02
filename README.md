# travis

Travis is a helper package for dealing with webhooks. It contains 2 functions
and 4 exported types. Everything is done by the functions `GetPayload` and `GetPayloadFromRequest`.

#### GetPayload(io.Reader) (*Payload, error)

This func will just parse the payload inside the reader without making any verifications

#### GetPayloadFromRequest(*http.Request) (*Payload, error)

This function will verify the integrity of the request and then parse the payload inside the body.
This is done by applying the steps listed in [here][1] and using the _official_ script referenced below
that section. The script can be found [here][2].

#### type Color int

This type provides colors to represents the state inside the webhook

```go
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
```

#### type Payload struct

This contains the data inside the travis payload, it also provides useful functions
to know more about the state of the payload such as:

* _Pending_ : returns true if a build has been requested
* _Passed_ : returns true if the build completed successfully
* _Fixed_ : returns true if the build completed successfully after a previously failed build
* _Broken_ : returns true if the build completed in failure after a previously successful build
* _Failed_ : returns true if the build is the first build for a new branch and has failed
* _StillFailing_ : returns true if the build completed in failure after a previously failed build
* _Canceled_ : returns true if the build was canceled
* _Errored_ : returns true if the build has errored
* _IsPullRequest_ : returns true if the event was caused by a pull request
* _IsPush_ : returns true if the event was caused by a push
* _IsCron_ : returns true if the event was caused by a cron
* _IsAPI_ : returns true if the event was caused by an api

```go
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
```

#### type Config struct

The type representing the `config` field inside the payload

```go
type Config struct {
	Sudo     bool   `json:"sudo,omitempty"`
	Dist     string `json:"dist,omitempty"`
	Language string `json:"language,omitempty"`
}
```

#### type Repository struct

The type representing the `repository` field inside the payload

```go
type Repository struct {
	ID        int64  `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`
	URL       string `json:"url,omitempty"`
}
```

[1]: https://docs.travis-ci.com/user/notifications/#Verifying-Webhook-requests
[2]: https://gist.github.com/theshapguy/7d10ea4fa39fab7db393021af959048e