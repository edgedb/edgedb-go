// This source file is part of the EdgeDB open source project.
//
// Copyright EdgeDB Inc. and the EdgeDB authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gel

import (
	"fmt"
	"math"
	"time"

	types "github.com/geldata/gel-go/internal/geltypes"
)

// Options for connecting to a Gel server
type Options struct {
	// Host is an Gel server host address, given as either an IP address or
	// domain name. (Unix-domain socket paths are not supported)
	//
	// Host cannot be specified alongside the 'dsn' argument, or
	// CredentialsFile option. Host will override all other credentials
	// resolved from any environment variables, or project credentials with
	// their defaults.
	Host string

	// Port is a port number to connect to at the server host.
	//
	// Port cannot be specified alongside the 'dsn' argument, or
	// CredentialsFile option. Port will override all other credentials
	// resolved from any environment variables, or project credentials with
	// their defaults.
	Port int

	// Credentials is a JSON string containing connection credentials.
	//
	// Credentials cannot be specified alongside the 'dsn' argument, Host,
	// Port, or CredentialsFile.  Credentials will override all other
	// credentials not present in the credentials string with their defaults.
	Credentials []byte

	// CredentialsFile is a path to a file containing connection credentials.
	//
	// CredentialsFile cannot be specified alongside the 'dsn' argument, Host,
	// Port, or Credentials.  CredentialsFile will override all other
	// credentials not present in the credentials file with their defaults.
	CredentialsFile string

	// User is the name of the database role used for authentication.
	//
	// If not specified, the value is resolved from any compound
	// argument/option, then from GEL_USER, then any compound environment
	// variable, then project credentials.
	User string

	// Database is the name of the database to connect to.
	//
	// If not specified, the value is resolved from any compound
	// argument/option, then from EDGEDB_DATABASE, then any compound
	// environment variable, then project credentials.
	//
	// Deprecated: Database has been replaced by Branch
	Database string

	// Branch is the name of the branch to use.
	//
	// If not specified, the value is resolved from any compound
	// argument/option, then from GEL_BRANCH, then any compound environment
	// variable, then project credentials.
	Branch string

	// Password to be used for authentication, if the server requires one.
	//
	// If not specified, the value is resolved from any compound
	// argument/option, then from GEL_PASSWORD, then any compound
	// environment variable, then project credentials.
	// Note that the use of the environment variable is discouraged
	// as other users and applications may be able to read it
	// without needing specific privileges.
	Password types.OptionalStr

	// ConnectTimeout is used when establishing connections in the background.
	ConnectTimeout time.Duration

	// WaitUntilAvailable determines how long to wait
	// to reestablish a connection.
	WaitUntilAvailable time.Duration

	// Concurrency determines the maximum number of connections.
	// If Concurrency is zero, max(4, runtime.NumCPU()) will be used.
	// Has no effect for single connections.
	Concurrency uint

	// Parameters used to configure TLS connections to Gel server.
	TLSOptions TLSOptions

	// Read the TLS certificate from this file.
	// DEPRECATED, use TLSOptions.CAFile instead.
	TLSCAFile string

	// Specifies how strict TLS validation is.
	// DEPRECATED, use TLSOptions.SecurityMode instead.
	TLSSecurity string

	// ServerSettings is currently unused.
	ServerSettings map[string][]byte

	// SecretKey is used to connect to cloud instances.
	SecretKey string

	// WarningHandler is invoked when Gel returns warnings. Defaults to
	// gel.LogWarnings.
	WarningHandler WarningHandler
}

// TLSOptions contains the parameters needed to configure TLS on Gel
// server connections.
type TLSOptions struct {
	// PEM-encoded CA certificate
	CA []byte
	// Path to a PEM-encoded CA certificate file
	CAFile string
	// Determines how strict we are with TLS checks
	SecurityMode TLSSecurityMode
	// Used to verify the hostname on the returned certificates
	ServerName string
}

// TLSSecurityMode specifies how strict TLS validation is.
type TLSSecurityMode string

const (
	// TLSModeDefault makes security mode inferred from other options
	TLSModeDefault TLSSecurityMode = "default"
	// TLSModeInsecure results in no certificate verification whatsoever
	TLSModeInsecure TLSSecurityMode = "insecure"
	// TLSModeNoHostVerification enables certificate verification
	// against CAs, but hostname matching is not performed.
	TLSModeNoHostVerification TLSSecurityMode = "no_host_verification"
	// TLSModeStrict enables full certificate and hostname verification.
	TLSModeStrict TLSSecurityMode = "strict"
)

// RetryBackoff returns the duration to wait after the nth attempt
// before making the next attempt when retrying a transaction.
type RetryBackoff func(n int) time.Duration

func defaultBackoff(attempt int) time.Duration {
	backoff := math.Pow(2.0, float64(attempt)) * 100.0
	jitter := rnd.Float64() * 100.0
	return time.Duration(backoff+jitter) * time.Millisecond
}

// RetryCondition represents scenarios that can cause a transaction
// run in Tx() methods to be retried.
type RetryCondition int

// The following conditions can be configured with a custom RetryRule.
// See RetryOptions.
const (
	// TxConflict indicates that the server could not complete a transaction
	// because it encountered a deadlock or serialization error.
	TxConflict = iota

	// NetworkError indicates that the transaction was interupted
	// by a network error.
	NetworkError
)

// NewRetryRule returns the default RetryRule value.
func NewRetryRule() RetryRule {
	return RetryRule{
		fromFactory: true,
		attempts:    3,
		backoff:     defaultBackoff,
	}
}

// RetryRule determines how transactions should be retried when run in Tx()
// methods. See Client.Tx() for details.
type RetryRule struct {
	// fromFactory indicates that a RetryOptions value was created using
	// NewRetryOptions() and not created directly. Requiring users to use the
	// factory function allows for nonzero default values.
	fromFactory bool

	// Total number of times to attempt a transaction.
	// attempts <= 0 indicate that a default value should be used.
	attempts int

	// backoff determines how long to wait between transaction attempts.
	// nil indicates that a default function should be used.
	backoff RetryBackoff
}

// WithAttempts sets the rule's attempts. attempts must be greater than zero.
func (r RetryRule) WithAttempts(attempts int) RetryRule {
	if attempts < 1 {
		panic(fmt.Sprintf(
			"RetryRule attempts must be greater than 0, got %v",
			attempts,
		))
	}

	r.attempts = attempts
	return r
}

// WithBackoff returns a copy of the RetryRule with backoff set to fn.
func (r RetryRule) WithBackoff(fn RetryBackoff) RetryRule {
	if fn == nil {
		panic("the backoff function must not be nil")
	}

	r.backoff = fn
	return r
}

// NewRetryOptions returns the default retry options.
func NewRetryOptions() RetryOptions {
	return RetryOptions{fromFactory: true}.WithDefault(NewRetryRule())
}

// RetryOptions configures how Tx() retries failed transactions.  Use
// NewRetryOptions to get a default RetryOptions value instead of creating one
// yourself.
type RetryOptions struct {
	fromFactory bool
	txConflict  RetryRule
	network     RetryRule
}

// WithDefault sets the rule for all conditions to rule.
func (o RetryOptions) WithDefault(rule RetryRule) RetryOptions { // nolint:gocritic,lll
	if !rule.fromFactory {
		panic("RetryRule not created with NewRetryRule() is not valid")
	}

	o.txConflict = rule
	o.network = rule
	return o
}

// WithCondition sets the retry rule for the specified condition.
func (o RetryOptions) WithCondition( // nolint:gocritic
	condition RetryCondition,
	rule RetryRule,
) RetryOptions {
	if !rule.fromFactory {
		panic("RetryRule not created with NewRetryRule() is not valid")
	}

	switch condition {
	case TxConflict:
		o.txConflict = rule
	case NetworkError:
		o.network = rule
	default:
		panic(fmt.Sprintf("unexpected condition: %v", condition))
	}

	return o
}

func (o RetryOptions) ruleForException(err Error) (RetryRule, error) {
	switch {
	case err.Category(TransactionConflictError):
		return o.txConflict, nil
	case err.Category(ClientError):
		return o.network, nil
	default:
		return RetryRule{}, &clientError{
			msg: fmt.Sprintf("unexpected error type: %T", err),
		}
	}
}

// IsolationLevel documentation can be found here
// https://www.edgedb.com/docs/reference/edgeql/tx_start#parameters
type IsolationLevel string

const (
	// Serializable is the only isolation level
	Serializable IsolationLevel = "serializable"
)

// NewTxOptions returns the default TxOptions value.
func NewTxOptions() TxOptions {
	return TxOptions{
		fromFactory: true,
		isolation:   Serializable,
	}
}

// TxOptions configures how transactions behave.
type TxOptions struct {
	// fromFactory indicates that a TxOptions value was created using
	// NewTxOptions() and not created directly with TxOptions{}.
	// Requiring users to use the factory function allows for nonzero
	// default values.
	fromFactory bool

	readOnly   bool
	deferrable bool
	isolation  IsolationLevel
}

// WithIsolation returns a copy of the TxOptions
// with the isolation level set to i.
func (o TxOptions) WithIsolation(i IsolationLevel) TxOptions {
	if i != Serializable {
		panic(fmt.Sprintf("unknown isolation level: %q", i))
	}

	o.isolation = i
	return o
}

// WithReadOnly returns a shallow copy of the client
// with the transaction read only access mode set to r.
func (o TxOptions) WithReadOnly(r bool) TxOptions {
	o.readOnly = r
	return o
}

// WithDeferrable returns a shallow copy of the client
// with the transaction deferrable mode set to d.
func (o TxOptions) WithDeferrable(d bool) TxOptions {
	o.deferrable = d
	return o
}

func (o TxOptions) startTxQuery() string { // nolint:gocritic
	query := "START TRANSACTION"

	switch o.isolation {
	case Serializable:
		query += " ISOLATION SERIALIZABLE"
	default:
		panic(fmt.Sprintf("unknown isolation level: %q", o.isolation))
	}

	if o.readOnly {
		query += ", READ ONLY"
	} else {
		query += ", READ WRITE"
	}

	if o.deferrable {
		query += ", DEFERRABLE"
	} else {
		query += ", NOT DEFERRABLE"
	}

	query += ";"
	return query
}

// WithTxOptions returns a shallow copy of the client
// with the TxOptions set to opts.
func (p Client) WithTxOptions(opts TxOptions) *Client { // nolint:gocritic
	if !opts.fromFactory {
		panic("TxOptions not created with NewTxOptions() are not valid")
	}

	p.txOpts = opts
	return &p
}

// WithRetryOptions returns a shallow copy of the client
// with the RetryOptions set to opts.
func (p Client) WithRetryOptions( // nolint:gocritic
	opts RetryOptions,
) *Client {
	if !opts.fromFactory {
		panic("RetryOptions not created with NewRetryOptions() are not valid")
	}

	p.retryOpts = opts
	return &p
}

// WithConfig sets configuration values for the returned client.
func (p Client) WithConfig( // nolint:gocritic
	cfg map[string]interface{},
) *Client {
	state := copyState(p.state)

	var config map[string]interface{}
	if c, ok := state["config"]; ok {
		config = c.(map[string]interface{})
	} else {
		config = make(map[string]interface{}, len(cfg))
	}

	for k, v := range cfg {
		config[k] = v
	}

	state["config"] = config
	p.state = state
	return &p
}

// WithoutConfig unsets configuration values for the returned client.
func (p Client) WithoutConfig(key ...string) *Client { // nolint:gocritic
	state := copyState(p.state)

	if c, ok := state["config"]; ok {
		config := c.(map[string]interface{})
		for _, k := range key {
			delete(config, k)
		}
	}

	p.state = state
	return &p
}

// ModuleAlias is an alias name and module name pair.
type ModuleAlias struct {
	Alias  string
	Module string
}

// WithModuleAliases sets module name aliases for the returned client.
func (p Client) WithModuleAliases( // nolint:gocritic
	aliases ...ModuleAlias,
) *Client {
	state := copyState(p.state)

	var a []interface{}
	if b, ok := state["aliases"]; ok {
		a = b.([]interface{})
	}

	for i := 0; i < len(aliases); i++ {
		a = append(a, []interface{}{aliases[i].Alias, aliases[i].Module})
	}

	state["aliases"] = a
	p.state = state
	return &p
}

// WithoutModuleAliases unsets module name aliases for the returned client.
func (p Client) WithoutModuleAliases( // nolint:gocritic
	aliases ...string,
) *Client {
	state := copyState(p.state)

	if a, ok := state["aliases"]; ok {
		blacklist := make(map[string]struct{}, len(aliases))
		for _, name := range aliases {
			blacklist[name] = struct{}{}
		}

		var without []interface{}
		for _, p := range a.([]interface{}) {
			pair := p.([]interface{})
			key := pair[0].(string)
			if _, ok := blacklist[key]; !ok {
				without = append(without, []interface{}{key, pair[1]})
			}
		}

		state["aliases"] = without
	}

	p.state = state
	return &p
}

// WithGlobals sets values for global variables for the returned client.
func (p Client) WithGlobals( // nolint:gocritic
	globals map[string]interface{},
) *Client {
	state := copyState(p.state)

	var g map[string]interface{}
	if x, ok := state["globals"]; ok {
		g = x.(map[string]interface{})
	} else {
		g = make(map[string]interface{}, len(globals))
	}

	for k, v := range globals {
		g[k] = v
	}

	state["globals"] = g
	p.state = state
	return &p
}

// WithoutGlobals unsets values for global variables for the returned client.
func (p Client) WithoutGlobals(globals ...string) *Client { // nolint:gocritic
	state := copyState(p.state)

	if c, ok := state["globals"]; ok {
		config := c.(map[string]interface{})
		for _, k := range globals {
			delete(config, k)
		}
	}

	p.state = state
	return &p
}

// WithWarningHandler sets the warning handler for the returned client. If
// warningHandler is nil gel.LogWarnings is used.
func (p Client) WithWarningHandler( // nolint:gocritic
	warningHandler WarningHandler,
) *Client {
	if warningHandler == nil {
		warningHandler = LogWarnings
	}

	p.warningHandler = warningHandler
	return &p
}
