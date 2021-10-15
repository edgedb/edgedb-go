// This source file is part of the EdgeDB open source project.
//
// Copyright 2020-present EdgeDB Inc. and the EdgeDB authors.
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

package edgedb

import (
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var errNoTOMLFound = errors.New("no edgedb.toml found")
var isDSNLike = regexp.MustCompile(`(?i)^[a-z]+://`)
var isIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

type connConfig struct {
	addr               dialArgs
	user               string
	password           string
	database           string
	connectTimeout     time.Duration
	waitUntilAvailable time.Duration
	serverSettings     map[string]string
	tlsConfig          *tls.Config
}

type dialArgs struct {
	network string
	address string
}

type cfgVal struct {
	val    interface{}
	source string
}

type configResolver struct {
	host              cfgVal // string
	port              cfgVal // int
	database          cfgVal // string
	user              cfgVal // string
	password          cfgVal // OptionalStr
	tlsCAData         cfgVal // []byte
	tlsVerifyHostname cfgVal // OptionalBool
	serverSettings    map[string]string
}

func (r *configResolver) setHost(val, source string) error {
	if r.host.val != nil {
		return nil
	}
	if strings.Contains(val, "/") {
		return fmt.Errorf(
			"invalid host: unix socket paths not supported, got %q", val)
	}
	if val == "" || strings.Contains(val, ",") {
		return fmt.Errorf(`invalid host: %q`, val)
	}
	r.host = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) setPort(val int, source string) error {
	if r.port.val != nil {
		return nil
	}
	if val < 1 || val > 65535 {
		return fmt.Errorf(`invalid port: %v`, val)
	}
	r.port = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) setPortStr(val, source string) error {
	if r.port.val != nil {
		return nil
	}
	port, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", val, err)
	}
	return r.setPort(port, source)
}

func (r *configResolver) setDatabase(val, source string) error {
	if r.database.val != nil {
		return nil
	}
	if val == "" {
		return errors.New(`invalid database name: ""`)
	}
	r.database = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) setUser(val, source string) error {
	if r.user.val != nil {
		return nil
	}
	if val == "" {
		return errors.New(`invalid user name: ""`)
	}
	r.user = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) setPassword(val, source string) {
	if r.password.val != nil {
		return
	}
	r.password = cfgVal{val: val, source: source}
}

func (r *configResolver) setTLSCAData(data []byte, source string) {
	if r.tlsCAData.val != nil {
		return
	}
	r.tlsCAData = cfgVal{val: data, source: source}
}

func (r *configResolver) setTLSCAFile(file, source string) error {
	if r.tlsCAData.val != nil {
		return nil
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	r.tlsCAData = cfgVal{val: data, source: source}
	return nil
}

func (r *configResolver) setTLSVerifyHostname(val bool, source string) {
	if r.tlsVerifyHostname.val != nil {
		return
	}
	r.tlsVerifyHostname = cfgVal{val: val, source: source}
}

func (r *configResolver) setTLSVerifyHostnameStr(val, source string) error {
	if r.tlsVerifyHostname.val != nil {
		return nil
	}
	v, err := parseVerifyHostname(val)
	if err != nil {
		return err
	}
	r.setTLSVerifyHostname(v, source)
	return nil
}

func (r *configResolver) addServerSettings(s map[string]string) {
	for k, v := range s {
		if _, ok := r.serverSettings[k]; !ok {
			r.serverSettings[k] = v
		}
	}
}

func (r *configResolver) resolveOptions(opts *Options) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid edgedb.Options: %w", err)
		}
	}()

	if opts.Host != "" {
		if e := r.setHost(opts.Host, "Host option"); e != nil {
			return e
		}
	}

	if opts.Port != 0 {
		if e := r.setPort(opts.Port, "Port option"); e != nil {
			return e
		}
	}

	if opts.Database != "" {
		if e := r.setDatabase(opts.Database, "Database options"); e != nil {
			return e
		}
	}

	if opts.User != "" {
		if e := r.setUser(opts.User, "User options"); e != nil {
			return e
		}
	}

	if pwd, ok := opts.Password.Get(); ok {
		r.setPassword(pwd, "Password option")
	}

	if opts.TLSCAFile != "" {
		if e := r.setTLSCAFile(opts.TLSCAFile, "TLSCAFile option"); e != nil {
			return e
		}
	}

	if val, ok := opts.TLSVerifyHostname.Get(); ok {
		r.setTLSVerifyHostname(val, "TLSVerifyHostname option")
	}
	r.addServerSettings(opts.ServerSettings)
	return nil
}

func (r *configResolver) resolveDSN(dsn, source string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid DSN: %w", err)
		}
	}()

	uri, query, err := parseDSN(dsn)
	if err != nil {
		return err
	}

	val, err := popDSNValue(query, uri.Hostname(), "host", r.host.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		if e := r.setHost(val.val.(string), source+val.source); e != nil {
			return e
		}
	}

	val, err = popDSNValue(query, uri.Port(), "port", r.port.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		if e := r.setPortStr(val.val.(string), source+val.source); e != nil {
			return e
		}
	}

	db := strings.TrimPrefix(uri.Path, "/")
	val, err = popDSNValue(query, db, "database", r.database.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		db := strings.TrimPrefix(val.val.(string), "/")
		if e := r.setDatabase(db, source+val.source); e != nil {
			return e
		}
	}

	val, err = popDSNValue(query, uri.User.Username(), "user",
		r.user.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		if e := r.setUser(val.val.(string), source+val.source); e != nil {
			return e
		}
	}

	pwd, ok := uri.User.Password()
	val, err = popDSNValue(query, "", "password", r.password.val == nil)
	if err != nil {
		return err
	}
	if r.password.val == nil && ok {
		// XXX: what is the source supposed to be here?
		r.setPassword(pwd, source)
	} else if r.password.val == nil && val.val != nil {
		r.setPassword(val.val.(string), source+val.source)
	}

	val, err = popDSNValue(query, "", "tls_cert_file", r.tlsCAData.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		if e := r.setTLSCAFile(val.val.(string), source+val.source); e != nil {
			return e
		}
	}

	val, err = popDSNValue(query, "", "tls_verify_hostname",
		r.tlsVerifyHostname.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		err = r.setTLSVerifyHostnameStr(val.val.(string), source+val.source)
		if err != nil {
			return err
		}
	}

	r.addServerSettings(query)
	return nil
}

func (r *configResolver) resolveCredentials(
	instance, credentials, source string,
) error {
	if instance != "" && credentials != "" {
		return errors.New("cannot have both instance name and " +
			"credentials file")
	}

	if instance != "" {
		if !isIdentifier.MatchString(instance) {
			return fmt.Errorf("invalid instance name %q", instance)
		}
		var err error
		credentials, err = findConfigPath("credentials", instance+".json")
		if err != nil {
			return err
		}
	}

	creds, err := readCredentials(credentials)
	if err != nil {
		if instance != "" {
			return fmt.Errorf(
				"cannot read credentials for instance %q: %w", instance, err)
		}
		return fmt.Errorf(
			"cannot read credentials for file %q: %w", credentials, err)
	}

	if host, ok := creds.host.Get(); ok && host != "" {
		if e := r.setHost(host, source); e != nil {
			return e
		}
	}

	if port, ok := creds.port.Get(); ok && port != 0 {
		if e := r.setPort(int(port), source); e != nil {
			return e
		}
	}

	if db, ok := creds.database.Get(); ok && db != "" {
		if e := r.setDatabase(db, source); e != nil {
			return e
		}
	}

	if e := r.setUser(creds.user, source); e != nil {
		return e
	}

	if pwd, ok := creds.password.Get(); ok {
		r.setPassword(pwd, source)
	}

	if data, ok := creds.certData.Get(); ok && len(data) > 0 {
		r.setTLSCAData(data, source)
	}

	if verifyHostname, ok := creds.verifyHostname.Get(); ok {
		r.setTLSVerifyHostname(verifyHostname, source)
	}

	return nil
}

func (r *configResolver) resolveEnvVars() (bool, error) {
	if db, ok := os.LookupEnv("EDGEDB_DATABASE"); ok {
		err := r.setDatabase(db, "EDGEDB_DATABASE environment variable")
		if err != nil {
			return false, err
		}
	}

	if user, ok := os.LookupEnv("EDGEDB_USER"); ok {
		err := r.setUser(user, "EDGEDB_USER environment variable")
		if err != nil {
			return false, err
		}
	}

	if pwd, ok := os.LookupEnv("EDGEDB_PASSWORD"); ok {
		r.setPassword(pwd, "EDGEDB_PASSWORD environment variable")
	}

	if file, ok := os.LookupEnv("EDGEDB_TLS_CA_FILE"); ok {
		e := r.setTLSCAFile(file, "EDGEDB_TLS_CA_FILE environment variable")
		if e != nil {
			return false, e
		}
	}

	if verify, ok := os.LookupEnv("EDGEDB_TLS_VERIFY_HOSTNAME"); ok {
		source := "EDGEDB_TLS_VERIFY_HOSTNAME environment variable"
		if e := r.setTLSVerifyHostnameStr(verify, source); e != nil {
			return false, e
		}
	}

	var names []string
	dsn, dsnOk := os.LookupEnv("EDGEDB_DSN")
	if dsnOk {
		names = append(names, "EDGEDB_DSN")
	}
	instance, instanceOk := os.LookupEnv("EDGEDB_INSTANCE")
	if instanceOk {
		names = append(names, "EDGEDB_INSTANCE")
	}
	credentials, credsOk := os.LookupEnv("EDGEDB_CREDENTIALS_FILE")
	if credsOk {
		names = append(names, "EDGEDB_CREDENTIALS_FILE")
	}
	host, hostOk := os.LookupEnv("EDGEDB_HOST")
	if hostOk {
		names = append(names, "EDGEDB_HOST")
	}
	port, portOk := os.LookupEnv("EDGEDB_PORT")
	if portOk && strings.HasPrefix(port, "tcp://") {
		// EDGEDB_PORT is set by 'docker --link' so ignore and warn
		log.Println(
			"Warning: ignoring EDGEDB_PORT in 'tcp://host:port' format")
		portOk = false
	}

	if !hostOk && portOk {
		names = append(names, "EDGEDB_PORT")
	}

	if len(names) > 1 {
		return false, fmt.Errorf(
			"mutually exclusive environment variables set: %v",
			strings.Join(names, ", "))
	}

	switch {
	case hostOk || portOk:
		if portOk {
			err := r.setPortStr(port, "EDGEDB_PORT environment variable")
			if err != nil {
				return false, err
			}
		}

		if hostOk {
			err := r.setHost(host, "EDGEDB_HOST environment variable")
			if err != nil {
				return false, err
			}
		}
	case dsnOk:
		e := r.resolveDSN(dsn, "EDGEDB_DSN environment variable")
		if e != nil {
			return false, e
		}
	case instanceOk || credsOk:
		source := "EDGEDB_CREDENTIALS_FILE environment variable"
		if instanceOk {
			source = "EDGEDB_INSTANCE environment variable"
		}
		err := r.resolveCredentials(instance, credentials, source)
		if err != nil {
			return false, err
		}
	default:
		return false, nil
	}

	return true, nil
}

func (r *configResolver) resolveTOML() (string, error) {
	toml, err := findEdgeDBTOML()
	if err != nil {
		return "", err
	}
	stashDir, err := stashPath(filepath.Dir(toml))
	if err != nil {
		return "", err
	}

	if _, e := os.Stat(stashDir); os.IsNotExist(e) {
		return "", errors.New("Found `edgedb.toml` " +
			"but the project is not initialized. Run `edgedb project init`.")
	}

	instance, err := ioutil.ReadFile(filepath.Join(stashDir, "instance-name"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(instance)), nil
}

func (r *configResolver) config(opts *Options) (*connConfig, error) {
	host := "localhost"
	if r.host.val != nil {
		host = r.host.val.(string)
	}

	port := 5656
	if r.port.val != nil {
		port = r.port.val.(int)
	}

	database := "edgedb"
	if r.database.val != nil {
		database = r.database.val.(string)
	}

	user := "edgedb"
	if r.user.val != nil {
		user = r.user.val.(string)
	}

	var certData []byte
	if r.tlsCAData.val != nil {
		certData = r.tlsCAData.val.([]byte)
	}

	var roots *x509.CertPool
	if len(certData) != 0 {
		roots = x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(certData); !ok {
			return nil, errors.New("invalid certificate data")
		}
	} else {
		var err error
		roots, err = getSystemCertPool()
		if err != nil {
			return nil, err
		}
	}

	verifyHostname := len(certData) == 0
	if r.tlsVerifyHostname.val != nil {
		verifyHostname = r.tlsVerifyHostname.val.(bool)
	}

	tlsConfig := &tls.Config{
		RootCAs:    roots,
		NextProtos: []string{"edgedb-binary"},
	}

	if os.Getenv("EDGEDB_INSECURE_DEV_MODE") != "" {
		tlsConfig.InsecureSkipVerify = true
	} else if !verifyHostname {
		// Set InsecureSkipVerify to skip the default validation we are
		// replacing. This will not disable VerifyConnection.
		tlsConfig.InsecureSkipVerify = true

		tlsConfig.VerifyConnection = func(cs tls.ConnectionState) error {
			opts := x509.VerifyOptions{
				Intermediates: x509.NewCertPool(),
				Roots:         roots,
			}
			for _, cert := range cs.PeerCertificates[1:] {
				opts.Intermediates.AddCert(cert)
			}
			_, err := cs.PeerCertificates[0].Verify(opts)
			return err
		}
	}

	waitUntilAvailable := opts.WaitUntilAvailable
	if waitUntilAvailable == 0 {
		waitUntilAvailable = 30 * time.Second
	}

	password := ""
	if r.password.val != nil {
		password = r.password.val.(string)
	}

	return &connConfig{
		addr:               dialArgs{"tcp", fmt.Sprintf("%v:%v", host, port)},
		user:               user,
		password:           password,
		database:           database,
		connectTimeout:     opts.ConnectTimeout,
		waitUntilAvailable: waitUntilAvailable,
		serverSettings:     r.serverSettings,
		tlsConfig:          tlsConfig,
	}, nil
}

func newConfigResolver(dsn string, opts *Options) (*configResolver, error) {
	cfg := &configResolver{serverSettings: map[string]string{}}

	var instance string
	if !isDSNLike.MatchString(dsn) {
		instance = dsn
		dsn = ""
	}

	var names []string
	if dsn != "" || instance != "" {
		names = append(names, "dsn")
	}
	if opts.CredentialsFile != "" {
		names = append(names, "edgedb.Options.CredentialsFile")
	}
	if opts.Host != "" {
		names = append(names, "edgedb.Options.Host")
	} else if opts.Port != 0 {
		names = append(names, "edgedb.Options.Port")
	}
	if len(names) > 1 {
		return nil, fmt.Errorf(
			"mutually exclusive connection options specified: %v",
			strings.Join(names, ", "))
	}

	if e := cfg.resolveOptions(opts); e != nil {
		return nil, e
	}

	switch {
	case opts.Host != "" || opts.Port != 0:
		// stop here since there is a host or port
	case dsn != "":
		if e := cfg.resolveDSN(dsn, "DSN option"); e != nil {
			return nil, e
		}
	case instance != "" || opts.CredentialsFile != "":
		source := "CredentialsFile option"
		if instance != "" {
			source = "dsn (parsed as instance name)"
		}
		err := cfg.resolveCredentials(instance, opts.CredentialsFile, source)
		if err != nil {
			return nil, err
		}
	default:
		ok, err := cfg.resolveEnvVars()
		if err != nil {
			return nil, err
		} else if ok {
			break
		}

		instance, err := cfg.resolveTOML()
		if errors.Is(err, errNoTOMLFound) {
			return nil, errors.New(
				"no `edgedb.toml` found and no connection options " +
					"specified either via arguments to connect API " +
					"or via environment variables " +
					"EDGEDB_HOST/EDGEDB_PORT, EDGEDB_INSTANCE, " +
					"EDGEDB_DSN or EDGEDB_CREDENTIALS_FILE",
			)
		}
		if err != nil {
			return nil, err
		}

		source := fmt.Sprintf("project linked instance (%q)", instance)
		err = cfg.resolveCredentials(instance, "", source)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func parseConnectDSNAndArgs(dsn string, opts *Options) (*connConfig, error) {
	resolver, err := newConfigResolver(dsn, opts)
	if err != nil {
		return nil, &configurationError{err: err}
	}

	c, err := resolver.config(opts)
	if err != nil {
		return nil, &configurationError{err: err}
	}

	return c, nil
}

func parseDSN(dsn string) (*url.URL, map[string]string, error) {
	uri, err := url.Parse(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("could not parse DSN %q: %w", dsn, err)
	}

	if uri.Scheme != "edgedb" {
		return nil, nil, fmt.Errorf(
			`scheme is expected to be "edgedb", got %q`, uri.Scheme)
	}

	query, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not parse DSN query parameters %q: %w", uri.RawQuery, err)
	}

	vals := make(map[string]string, len(query))
	for k, v := range query {
		if len(v) > 1 {
			return nil, nil, fmt.Errorf(
				`duplicate query parameter %q in %v`, k, dsn)
		}
		vals[k] = v[0]
	}

	if e := validateQueryArg(vals, "host", uri.Hostname()); e != nil {
		return nil, nil, e
	}

	if e := validateQueryArg(vals, "port", uri.Port()); e != nil {
		return nil, nil, e
	}

	db := strings.TrimPrefix(uri.Path, "/")
	if e := validateQueryArg(vals, "database", db); e != nil {
		return nil, nil, e
	}

	if e := validateQueryArg(vals, "user", uri.User.Username()); e != nil {
		return nil, nil, e
	}

	pwd, ok := uri.User.Password()
	if ok {
		pwd = "non empty string"
	}
	if e := validateQueryArg(vals, "password", pwd); e != nil {
		return nil, nil, e
	}

	if e := validateQueryArg(vals, "tls_cert_file", ""); e != nil {
		return nil, nil, e
	}

	if e := validateQueryArg(vals, "tls_verify_hostname", ""); e != nil {
		return nil, nil, e
	}

	return uri, vals, nil
}

var dsnKeyLookup = map[string][]string{
	"host":          {"host", "host_env", "host_file"},
	"port":          {"port", "port_env", "port_file"},
	"database":      {"database", "database_env", "database_file"},
	"user":          {"user", "user_env", "user_file"},
	"password":      {"password", "password_env", "password_file"},
	"tls_cert_file": {"tls_cert_file", "tls_cert_file_env"},
	"tls_verify_hostname": {
		"tls_verify_hostname",
		"tls_verify_hostname_env",
		"tls_verify_hostname_file",
	},
}

func validateQueryArg(query map[string]string, name string, val string) error {
	var msgs []string
	if val != "" {
		msgs = append(msgs, fmt.Sprintf(`%v, `, name))
	}

	for _, name := range dsnKeyLookup[name] {
		_, ok := query[name]
		if ok {
			msgs = append(msgs, fmt.Sprintf("?%v=", name))
		}
	}

	if len(msgs) > 1 {
		return fmt.Errorf(
			"mutually exclusive query arguments specified: %v",
			strings.Join(msgs, " "))
	}

	return nil
}

func popDSNValue(
	query map[string]string,
	val string,
	name string,
	resolve bool,
) (cfgVal, error) {
	if val != "" {
		// XXX: what is the source supposed to be here?
		return cfgVal{val: val, source: "x_x"}, nil
	}

	var key string
	for _, k := range dsnKeyLookup[name] {
		if _, ok := query[k]; ok {
			key = k
			break
		}
	}

	val, ok := query[key]
	delete(query, key)

	if !resolve {
		return cfgVal{}, nil
	}

	switch {
	case ok && key == "tls_cert_file":
		source := fmt.Sprintf(" (%v: %q)", key, val)
		return cfgVal{val: val, source: source}, nil
	case ok && strings.HasSuffix(key, "_env"):
		v, k := os.LookupEnv(val)
		if !k {
			return cfgVal{}, fmt.Errorf(
				"%v environment variable %q is not set", key, val)
		}
		source := fmt.Sprintf(" (%v: %q)", key, val)
		return cfgVal{val: v, source: source}, nil
	case ok && strings.HasSuffix(key, "_file"):
		data, err := ioutil.ReadFile(val)
		if err != nil {
			return cfgVal{}, fmt.Errorf(
				"failed to read %v file %q: %w", key, val, err)
		}
		source := fmt.Sprintf(" (%v: %q)", key, val)
		return cfgVal{val: string(data), source: source}, nil
	case ok:
		source := fmt.Sprintf(" (%v: %q)", key, val)
		return cfgVal{val: val, source: source}, nil
	default:
		return cfgVal{}, nil
	}
}

func stashPath(p string) (string, error) {
	p, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" && !strings.HasPrefix(p, `\\`) {
		p = `\\?\` + p
	}

	hash := fmt.Sprintf("%x", sha1.Sum([]byte(p)))
	baseName := filepath.Base(p)
	dirName := baseName + "-" + hash

	return findConfigPath("projects", dirName)
}

func parseVerifyHostname(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "t", "yes", "y", "1", "on":
		return true, nil
	case "false", "f", "no", "n", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf(
			"tls_verify_hostname can only be one of yes/no, got %q", s)
	}
}

func oldConfigDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, ".edgedb"), nil
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func findConfigPath(suffix ...string) (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}

	parts := append([]string{dir}, suffix...)
	dir = filepath.Join(parts...)
	if exists(dir) {
		return dir, nil
	}

	fallback, err := oldConfigDir()
	if err != nil {
		return "", err
	}

	parts = append([]string{fallback}, suffix...)
	fallback = filepath.Join(parts...)

	if exists(fallback) {
		return fallback, nil
	}

	return dir, nil
}

func findEdgeDBTOML() (string, error) {
	// If the current directory can be reached via multiple paths (due to
	// symbolic links), Getwd may return any one of them.
	dir, err := os.Getwd()
	if err != nil {
		return "", &clientConnectionError{err: err}
	}

	dev, err := device(dir)
	if err != nil {
		return "", err
	}

	for {
		tomlPath := filepath.Join(dir, "edgedb.toml")
		if _, e := os.Stat(tomlPath); os.IsNotExist(e) {
			parent := filepath.Dir(dir)
			// Stop searching when dir is the root directory.
			if parent == dir {
				return "", errNoTOMLFound
			}

			pDev, err := device(parent)
			if err != nil {
				return "", fmt.Errorf(
					"searching for edgedb.toml in or above %q: %w",
					filepath.Dir(tomlPath), err)
			}

			// Stop searching at file system boundaries.
			if pDev != dev {
				return "", fmt.Errorf("%w: stopped searching for edgedb.toml "+
					"at file system boundary %q", err, dir)
			}

			dir = parent
			dev = pDev
			continue
		}
		return tomlPath, nil
	}
}
