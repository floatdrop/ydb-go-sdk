package credentials

import (
	"context"
	"errors"
)

var (
	// ErrCredentialsNoCredentials may be returned by Credentials implementations to
	// make driver act as if there no Credentials at all. That is, driver will
	// not send any token meta information during request.
	ErrCredentialsNoCredentials = errors.New("ydb: credentials: no credentials")
)

// Credentials is an interface that contains options used to authorize a
// client.
type Credentials interface {
	Token(context.Context) (string, error)
}

// Func is an adapter to allow the use of ordinary functions as
// Credentials.
type Func func(context.Context) (string, error)

// Token implements Credentials.
func (f Func) Token(ctx context.Context) (string, error) {
	return f(ctx)
}

// Token implements Credentials.
func (f Func) String() string {
	return "Func"
}

// AccessTokenCredentials implements Credentials interface with static
// authorization parameters.
type AccessTokenCredentials struct {
	AccessToken string

	sourceInfo string
}

func NewAccessTokenCredentials(AccessToken string, sourceInfo string) *AccessTokenCredentials {
	return &AccessTokenCredentials{
		AccessToken: AccessToken,
		sourceInfo:  sourceInfo,
	}
}

// Token implements Credentials.
func (a AccessTokenCredentials) Token(_ context.Context) (string, error) {
	return a.AccessToken, nil
}

// Token implements Credentials.
func (a AccessTokenCredentials) String() string {
	if a.sourceInfo == "" {
		return "AccessTokenCredentials"
	}
	return "AccessTokenCredentials created from " + a.sourceInfo
}

// anonymousCredentials implements Credentials interface with anonymous access
type anonymousCredentials struct {
	sourceInfo string
}

func NewAnonymousCredentials(sourceInfo string) *anonymousCredentials {
	return &anonymousCredentials{
		sourceInfo: sourceInfo,
	}
}

// Token implements Credentials.
func (a anonymousCredentials) Token(_ context.Context) (string, error) {
	return "", nil
}

// Token implements Credentials.
func (a anonymousCredentials) String() string {
	if a.sourceInfo == "" {
		return "anonymousCredentials"
	}
	return "anonymousCredentials created from " + a.sourceInfo
}

type multiCredentials struct {
	cs []Credentials
}

func (m *multiCredentials) Token(ctx context.Context) (token string, err error) {
	for _, c := range m.cs {
		token, err = c.Token(ctx)
		if err == nil {
			return
		}
	}
	if err == nil {
		err = ErrCredentialsNoCredentials
	}
	return
}

// MultiCredentials creates Credentials which represents multiple ways of
// obtaining token.
// Its Token() method proxies call to the underlying credentials in order.
// When first successful call met, it returns. If there are no successful
// calls, it returns last error.
func MultiCredentials(cs ...Credentials) Credentials {
	all := make([]Credentials, 0, len(cs))
	for _, c := range cs {
		if m, ok := c.(*multiCredentials); ok {
			all = append(all, m.cs...)
		} else {
			all = append(all, c)
		}
	}
	return &multiCredentials{
		cs: all,
	}
}
