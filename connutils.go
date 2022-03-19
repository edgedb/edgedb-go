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

	"github.com/edgedb/edgedb-go/internal/snc"
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
	tlsCAData          []byte
	tlsSecurity        string
	serverSettings     *snc.ServerSettings
}

func (c *connConfig) tlsConfig() (*tls.Config, error) {
	var roots *x509.CertPool
	if len(c.tlsCAData) != 0 {
		roots = x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(c.tlsCAData); !ok {
			return nil, errors.New("invalid certificate data")
		}
	} else {
		var err error
		roots, err = getSystemCertPool()
		if err != nil {
			return nil, err
		}
	}

	tlsConfig := &tls.Config{
		RootCAs:    roots,
		NextProtos: []string{"edgedb-binary"},
	}

	switch c.tlsSecurity {
	case "insecure_dev_mode", "insecure":
		tlsConfig.InsecureSkipVerify = true
	case "no_host_verification":
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

	return tlsConfig, nil
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
	host           cfgVal // string
	port           cfgVal // int
	database       cfgVal // string
	user           cfgVal // string
	password       cfgVal // OptionalStr
	tlsCAData      cfgVal // []byte
	tlsSecurity    cfgVal // string
	serverSettings *snc.ServerSettings
}

func (r *configResolver) setHost(val, source string) error {
	if r.host.val != nil {
		return nil
	}
	if strings.Contains(val, "/") {
		// unix socket
	} else if val == "" || strings.Contains(val, ",") {
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

func (r *configResolver) setTLSSecurity(val string, source string) error {
	if r.tlsSecurity.val != nil {
		return nil
	}

	switch val {
	case "insecure", "no_host_verification", "strict", "default":
	default:
		return invalidTLSSecurity(val)
	}
	r.tlsSecurity = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) addServerSettings(s map[string][]byte) {
	for k, v := range s {
		if _, ok := r.serverSettings.GetOk(k); !ok {
			r.serverSettings.Set(k, v)
		}
	}
}

func (r *configResolver) addServerSettingsStr(s map[string]string) {
	for k, v := range s {
		if _, ok := r.serverSettings.GetOk(k); !ok {
			r.serverSettings.Set(k, []byte(v))
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

	var caSources []string

	if opts.TLSCAFile != "" {
		caSources = append(caSources, "TLSCAFile")
		if e := r.setTLSCAFile(opts.TLSCAFile, "TLSCAFile option"); e != nil {
			return e
		}
	}

	if opts.TLSOptions.CA != nil {
		caSources = append(caSources, "TLSOptions.CA")
		r.setTLSCAData(opts.TLSOptions.CA, "TLSOptions.CA option")
	}

	if opts.TLSOptions.CAFile != "" {
		caSources = append(caSources, "TLSOptions.CAFile")
		if e := r.setTLSCAFile(
			opts.TLSOptions.CAFile, "TLSOptions.CAFile option"); e != nil {
			return e
		}
	}

	if len(caSources) > 1 {
		return fmt.Errorf(
			"mutually exclusive options set in Options: %v",
			englishList(caSources, "and"))
	}

	var secSources []string

	if opts.TLSSecurity != "" {
		secSources = append(secSources, "TLSSecurity")
		err = r.setTLSSecurity(opts.TLSSecurity, "TLSSecurity option")
		if err != nil {
			return err
		}
	}

	if opts.TLSOptions.SecurityMode != "" {
		secSources = append(secSources, "TLSOptions.SecurityMode")
		err = r.setTLSSecurity(
			string(opts.TLSOptions.SecurityMode),
			"TLSOptions.SecurityMode option")
	}

	if len(secSources) > 1 {
		return fmt.Errorf(
			"mutually exclusive options set in Options: %v",
			englishList(secSources, "and"))
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

	val, err = popDSNValue(query, "", "tls_ca_file", r.tlsCAData.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		if e := r.setTLSCAFile(val.val.(string), source+val.source); e != nil {
			return e
		}
	}

	val, err = popDSNValue(query, "", "tls_verify_hostname",
		r.tlsSecurity.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		switch val.val.(string) {
		case "insecure", "no_host_verification", "strict":
			err = r.setTLSSecurity(val.val.(string), source+val.source)
			if err != nil {
				return err
			}
		default:
			return invalidTLSSecurity(val.val.(string))
		}
	}

	val, err = popDSNValue(query, "", "tls_security", r.tlsSecurity.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		err = r.setTLSSecurity(val.val.(string), source+val.source)
		if err != nil {
			return err
		}
	}

	r.addServerSettingsStr(query)
	return nil
}

func (r *configResolver) resolveCredentials(
	instance, credentials, source string, paths *cfgPaths,
) error {
	if instance != "" && credentials != "" {
		return errors.New("cannot have both instance name and " +
			"credentials file")
	}

	if instance != "" {
		if !isIdentifier.MatchString(instance) {
			return fmt.Errorf("invalid instance name %q", instance)
		}
		dir, err := paths.CfgDir()
		if err != nil {
			return err
		}
		credentials = filepath.Join(dir, "credentials", instance+".json")
	}

	creds, err := readCredentials(credentials)
	if err != nil {
		if instance != "" {
			return fmt.Errorf(
				"cannot read credentials for instance %q: %w", instance, err)
		}
		return err
	}

	return r.applyCredentials(creds, source)
}

func (r *configResolver) applyCredentials(
	creds *credentials, source string,
) error {
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

	if data, ok := creds.ca.Get(); ok && len(data) > 0 {
		r.setTLSCAData(data, source)
	}

	if security, ok := creds.tlsSecurity.Get(); ok {
		if e := r.setTLSSecurity(security, source); e != nil {
			return e
		}
	}

	return nil
}

func (r *configResolver) resolveEnvVars(paths *cfgPaths) (bool, error) {
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

	var tlsCaSources []string

	if caString, ok := os.LookupEnv("EDGEDB_TLS_CA"); ok {
		r.setTLSCAData([]byte(caString), "EDGEDB_TLS_CA environment variable")
		tlsCaSources = append(tlsCaSources, "EDGEDB_TLS_CA")
	}

	if file, ok := os.LookupEnv("EDGEDB_TLS_CA_FILE"); ok {
		e := r.setTLSCAFile(file, "EDGEDB_TLS_CA_FILE environment variable")
		tlsCaSources = append(tlsCaSources, "EDGEDB_TLS_CA_FILE")
		if e != nil {
			return false, e
		}
	}

	if len(tlsCaSources) > 1 {
		return false, fmt.Errorf(
			"mutually exclusive environment variables set: %v",
			englishList(tlsCaSources, "and"))
	}

	if verify, ok := os.LookupEnv("EDGEDB_CLIENT_TLS_SECURITY"); ok {
		source := "EDGEDB_CLIENT_TLS_SECURITY environment variable"
		if e := r.setTLSSecurity(verify, source); e != nil {
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
			englishList(names, "and"))
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
		err := r.resolveCredentials(instance, credentials, source, paths)
		if err != nil {
			return false, err
		}
	default:
		return false, nil
	}

	return true, nil
}

func (r *configResolver) resolveTOML(paths *cfgPaths) (string, error) {
	toml, err := findEdgeDBTOML(paths)
	if err != nil {
		return "", err
	}

	stashDir, err := stashPath(filepath.Dir(toml), paths)
	if err != nil {
		return "", err
	}

	if !exists(stashDir) {
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

	tlsSecurity := "default"
	if r.tlsSecurity.val != nil {
		tlsSecurity = r.tlsSecurity.val.(string)
	}

	security, err := getEnvVarSetting("EDGEDB_CLIENT_SECURITY", "default",
		"default", "insecure_dev_mode", "strict")
	if err != nil {
		return nil, err
	}

	switch security {
	case "default":
	case "insecure_dev_mode":
		if tlsSecurity == "default" {
			tlsSecurity = "insecure"
		}
	case "strict":
		switch tlsSecurity {
		case "default":
			tlsSecurity = "strict"
		case "no_host_verification", "insecure":
			return nil, fmt.Errorf(
				"EDGEDB_CLIENT_SECURITY=strict but tls_security=%v, "+
					"tls_security must be set to strict "+
					"when EDGEDB_CLIENT_SECURITY is strict", tlsSecurity)
		}
	}

	if tlsSecurity == "default" {
		if len(certData) == 0 {
			tlsSecurity = "strict"
		} else {
			tlsSecurity = "no_host_verification"
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

	var addr dialArgs
	if strings.Contains(host, "/") {
		addr = dialArgs{network: "unix", address: host}
	} else {
		addr = dialArgs{
			network: "tcp", address: fmt.Sprintf("%v:%v", host, port),
		}
	}

	return &connConfig{
		addr:               addr,
		user:               user,
		password:           password,
		database:           database,
		connectTimeout:     opts.ConnectTimeout,
		waitUntilAvailable: waitUntilAvailable,
		serverSettings:     r.serverSettings,
		tlsCAData:          certData,
		tlsSecurity:        tlsSecurity,
	}, nil
}

func getEnvVarSetting(name, defalt string, values ...string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "default" || value == "" {
		return defalt, nil
	}

	for _, v := range values {
		if value == v {
			return value, nil
		}
	}

	return "", fmt.Errorf(
		"environment variable %v should be one of %v, got: %q",
		name, englishList(append(values, "default"), "or"), value)
}

func englishList(items []string, conjunction string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return strings.Join(items, fmt.Sprintf(" %v ", conjunction))
	default:
		last := len(items) - 1
		list := strings.Join(items[:last], ", ")
		return fmt.Sprintf("%v %v %v", list, conjunction, items[last])
	}
}

func newConfigResolver(
	dsn string,
	opts *Options,
	paths *cfgPaths,
) (*configResolver, error) {
	cfg := &configResolver{serverSettings: snc.NewServerSettings()}

	var instance string
	if !isDSNLike.MatchString(dsn) {
		instance = dsn
		dsn = ""
	}

	var names []string
	if dsn != "" || instance != "" {
		names = append(names, "dsn")
	}
	if opts.Credentials != nil {
		names = append(names, "edgedb.Options.Credentials")
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
			englishList(names, "and"))
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
		err := cfg.resolveCredentials(
			instance, opts.CredentialsFile, source, paths)
		if err != nil {
			return nil, err
		}
	case opts.Credentials != nil:
		source := "Credentials option"
		creds, err := parseCredentials(opts.Credentials, source)
		if err != nil {
			return nil, err
		}
		err = cfg.applyCredentials(creds, source)
		if err != nil {
			return nil, err
		}
	default:
		ok, err := cfg.resolveEnvVars(paths)
		if err != nil {
			return nil, err
		} else if ok {
			break
		}

		instance, err := cfg.resolveTOML(paths)
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
		err = cfg.resolveCredentials(instance, "", source, paths)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func parseConnectDSNAndArgs(
	dsn string,
	opts *Options,
	paths *cfgPaths,
) (*connConfig, error) {
	resolver, err := newConfigResolver(dsn, opts, paths)
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

	if e := validateQueryArg(vals, "tls_ca_file", ""); e != nil {
		return nil, nil, e
	}

	if e := validateQueryArg(vals, "tls_verify_hostname", ""); e != nil {
		return nil, nil, e
	}

	if e := validateQueryArg(vals, "tls_security", ""); e != nil {
		return nil, nil, e
	}

	return uri, vals, nil
}

var dsnKeyLookup = map[string][]string{
	"host":         {"host", "host_env", "host_file"},
	"port":         {"port", "port_env", "port_file"},
	"database":     {"database", "database_env", "database_file"},
	"user":         {"user", "user_env", "user_file"},
	"password":     {"password", "password_env", "password_file"},
	"tls_ca_file":  {"tls_ca_file", "tls_ca_file_env"},
	"tls_security": {"tls_security", "tls_security_env", "tls_security_file"},
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
			englishList(msgs, "and"))
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
	case ok && key == "tls_ca_file":
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

func stashPath(p string, paths *cfgPaths) (string, error) {
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

	cfgDir, err := paths.CfgDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cfgDir, "projects", dirName), nil
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

func cfgDir() (string, error) {
	dir, err := configDirOSSpecific()
	if err != nil {
		return "", err
	}

	if exists(dir) {
		return dir, nil
	}

	fallback, err := oldConfigDir()
	if err != nil {
		return "", err
	}

	if exists(fallback) {
		return fallback, err
	}

	return dir, nil
}

func newCfgPaths() *cfgPaths {
	paths := &cfgPaths{}
	paths.cwd, paths.cwdErr = os.Getwd()
	paths.cfgDir, paths.cfgDirErr = cfgDir()
	return paths
}

type cfgPaths struct {
	cwd       string
	cwdErr    error
	cfgDir    string
	cfgDirErr error
}

func (c *cfgPaths) Cwd() (string, error) { return c.cwd, c.cwdErr }

func (c *cfgPaths) CfgDir() (string, error) { return c.cfgDir, c.cfgDirErr }

func findEdgeDBTOML(paths *cfgPaths) (string, error) {
	// If the current directory can be reached via multiple paths (due to
	// symbolic links), Getwd may return any one of them.
	dir, err := paths.Cwd()
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
