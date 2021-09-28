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

type connConfig struct {
	addrs              []*dialArgs
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

type resolvedParams struct {
	host                    string
	hostSource              string
	port                    int
	portSource              string
	database                string
	databaseSource          string
	user                    string
	userSource              string
	password                string
	passwordSource          string
	passwordSet             bool
	tlsCAData               []byte
	tlsCADataSource         string
	tlsVerifyHostname       bool
	tlsVerifyHostnameSource string
	tlsVerifyHostnameSet    bool
	serverSettings          map[string]string
}

func (params *resolvedParams) setHost(val string, source string) error {
	if params.host != "" {
		return nil
	}
	if err := validateHost(val); err != nil {
		return err
	}
	params.host = val
	params.hostSource = source
	return nil
}

func (params *resolvedParams) setPort(val int, source string) error {
	if params.port != 0 {
		return nil
	}
	if err := validatePort(val); err != nil {
		return err
	}
	params.port = val
	params.portSource = source
	return nil
}

func (params *resolvedParams) setPortStr(val string, source string) error {
	if params.port != 0 {
		return nil
	}
	port, err := strconv.Atoi(val)
	if err != nil {
		return &configurationError{msg: fmt.Sprintf(
			"invalid port %q: %v",
			val, err,
		)}
	}
	return params.setPort(port, source)
}

func (params *resolvedParams) setDatabase(val string, source string) error {
	if params.database != "" {
		return nil
	}
	if val == "" {
		return &configurationError{msg: "invalid database name"}
	}
	params.database = val
	params.databaseSource = source
	return nil
}

func (params *resolvedParams) setUser(val string, source string) error {
	if params.user != "" {
		return nil
	}
	if val == "" {
		return &configurationError{msg: "invalid user name"}
	}
	params.user = val
	params.userSource = source
	return nil
}

func (params *resolvedParams) setPassword(val string, source string) {
	if params.passwordSet {
		return
	}
	params.password = val
	params.passwordSource = source
	params.passwordSet = true
}

func (params *resolvedParams) setTLSCAData(data []byte, source string) {
	if len(params.tlsCAData) != 0 {
		return
	}
	params.tlsCAData = data
	params.tlsCADataSource = source
}

func (params *resolvedParams) setTLSCAFile(file string, source string) error {
	if len(params.tlsCAData) != 0 {
		return nil
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return &configurationError{err: err}
	}
	params.tlsCAData = data
	params.tlsCADataSource = source
	return nil
}

func (params *resolvedParams) setTLSVerifyHostname(val bool, source string) {
	if params.tlsVerifyHostnameSet {
		return
	}
	params.tlsVerifyHostname = val
	params.tlsVerifyHostnameSource = source
	params.tlsVerifyHostnameSet = true
}

func (params *resolvedParams) setTLSVerifyHostnameStr(
	val string,
	source string,
) error {
	if params.tlsVerifyHostnameSet {
		return nil
	}
	verifyHostname, err := parseVerifyHostname(val)
	if err != nil {
		return err
	}
	params.setTLSVerifyHostname(verifyHostname, source)
	return nil
}

func (params *resolvedParams) addServerSettings(
	serverSettings map[string]string,
) {
	for k, v := range serverSettings {
		if _, keyExists := params.serverSettings[k]; !keyExists {
			params.serverSettings[k] = v
		}
	}
}

func validateHost(host string) error {
	if strings.Contains(host, "/") {
		return &configurationError{
			msg: "unix socket paths not supported",
		}
	}
	if host == "" || strings.Contains(host, ",") {
		return &configurationError{
			msg: fmt.Sprintf(`invalid host: %q`, host),
		}
	}
	return nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return &configurationError{
			msg: fmt.Sprintf(`invalid port: %v`, port),
		}
	}
	return nil
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

var isDSNLike = regexp.MustCompile(`(?i)^[a-z]+://`)
var isIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z_0-9]*$`)

func parseConnectDSNAndArgs(
	dsn string,
	opts *Options,
) (*connConfig, error) {
	resolvedConfig := resolvedParams{
		serverSettings: map[string]string{},
	}

	var instanceName string
	if !isDSNLike.MatchString(dsn) {
		instanceName = dsn
		dsn = ""
	}

	compoundOptionsCount := 0

	if opts.Database != "" {
		err := resolvedConfig.setDatabase(opts.Database, "Database option")
		if err != nil {
			return nil, err
		}
	}
	if opts.User != "" {
		err := resolvedConfig.setUser(opts.User, "User option")
		if err != nil {
			return nil, err
		}
	}
	if password, passwordIsSet := opts.Password.Get(); passwordIsSet {
		resolvedConfig.setPassword(password, "Password option")
	}
	if opts.TLSCAFile != "" {
		err := resolvedConfig.setTLSCAFile(opts.TLSCAFile, "TLSCAFile option")
		if err != nil {
			return nil, err
		}
	}
	if verifyHostname, verifyHostnameIsSet :=
		opts.TLSVerifyHostname.Get(); verifyHostnameIsSet {
		resolvedConfig.setTLSVerifyHostname(
			verifyHostname, "TLSVerifyHostname option",
		)
	}
	resolvedConfig.addServerSettings(opts.ServerSettings)

	if dsn != "" {
		compoundOptionsCount++
	}
	if instanceName != "" {
		compoundOptionsCount++
	}
	if opts.CredentialsFile != "" {
		compoundOptionsCount++
	}
	if opts.Host != "" {
		compoundOptionsCount++
	}
	if opts.Host == "" && opts.Port != 0 {
		compoundOptionsCount++
	}

	if compoundOptionsCount > 1 {
		return nil, &configurationError{
			msg: "Cannot have more than one of the following connection " +
				"options: dsn, CredentialsFile, or Host/Port"}
	}

	if compoundOptionsCount == 1 {
		if dsn != "" || opts.Host != "" || opts.Port != 0 {
			dsnSource := "dsn option"
			if dsn == "" {
				if opts.Port != 0 {
					err := resolvedConfig.setPort(opts.Port, "Port option")
					if err != nil {
						return nil, err
					}
				}
				if opts.Host != "" {
					err := validateHost(opts.Host)
					if err != nil {
						return nil, err
					}
					dsn = "edgedb://" + opts.Host
					dsnSource = "Host option"
				} else {
					dsn = "edgedb://"
					dsnSource = "Port option"
				}
			}

			err := parseDSNIntoConfig(&resolvedConfig, dsn, dsnSource)
			if err != nil {
				return nil, err
			}
		} else {
			err := loadCredentialsIntoConfig(
				&resolvedConfig,
				instanceName,
				instanceName != "",
				"dsn (parsed as instance name)",
				opts.CredentialsFile,
				opts.CredentialsFile != "",
				"CredentialsFile option",
			)
			if err != nil {
				return nil, err
			}
		}
	}

	if compoundOptionsCount == 0 {
		portEnv, portEnvExists := os.LookupEnv("EDGEDB_PORT")
		if portEnvExists && resolvedConfig.port == 0 &&
			strings.HasPrefix(portEnv, "tcp://") {
			// EDGEDB_PORT is set by 'docker --link' so ignore and warn
			log.Println(
				"Warning: EDGEDB_PORT in 'tcp://host:port' format, " +
					"so will be ignored",
			)
			portEnvExists = false
		}

		if database, dbEnvExists :=
			os.LookupEnv("EDGEDB_DATABASE"); dbEnvExists {
			err := resolvedConfig.setDatabase(
				database, "EDGEDB_DATABASE environment variable",
			)
			if err != nil {
				return nil, err
			}
		}
		if user, userEnvExists := os.LookupEnv("EDGEDB_USER"); userEnvExists {
			err := resolvedConfig.setUser(
				user, "EDGEDB_USER environment variable",
			)
			if err != nil {
				return nil, err
			}
		}
		if password, passwordEnvExists :=
			os.LookupEnv("EDGEDB_PASSWORD"); passwordEnvExists {
			resolvedConfig.setPassword(
				password, "EDGEDB_PASSWORD environment variable",
			)
		}
		if tlsCAFile, tlsCAFileEnvExists :=
			os.LookupEnv("EDGEDB_TLS_CA_FILE"); tlsCAFileEnvExists {
			err := resolvedConfig.setTLSCAFile(
				tlsCAFile, "EDGEDB_TLS_CA_FILE environment variable",
			)
			if err != nil {
				return nil, err
			}
		}
		if tlsVerifyHostname, tlsVerifyHostnameEnvExists :=
			os.LookupEnv(
				"EDGEDB_TLS_VERIFY_HOSTNAME"); tlsVerifyHostnameEnvExists {
			err := resolvedConfig.setTLSVerifyHostnameStr(
				tlsVerifyHostname, "EDGEDB_USER environment variable",
			)
			if err != nil {
				return nil, err
			}
		}

		dsn, dsnEnvExists := os.LookupEnv("EDGEDB_DSN")
		if dsnEnvExists {
			compoundOptionsCount++
		}
		instanceName, instanceEnvExists := os.LookupEnv("EDGEDB_INSTANCE")
		if instanceEnvExists {
			compoundOptionsCount++
		}
		credentialsFile, credFileEnvExists :=
			os.LookupEnv("EDGEDB_CREDENTIALS_FILE")
		if credFileEnvExists {
			compoundOptionsCount++
		}
		host, hostEnvExists := os.LookupEnv("EDGEDB_HOST")
		if hostEnvExists {
			compoundOptionsCount++
		}
		if !hostEnvExists && portEnvExists {
			compoundOptionsCount++
		}

		if compoundOptionsCount > 1 {
			return nil, &configurationError{
				msg: "Cannot have more than one of the following " +
					"environment variables: EDGEDB_DSN, EDGEDB_INSTANCE, " +
					"EDGEDB_CREDENTIALS_FILE, or EDGEDB_HOST/EDGEDB_PORT",
			}
		}

		if compoundOptionsCount == 1 {
			if dsnEnvExists || hostEnvExists || portEnvExists {
				dsnSource := "EDGEDB_DSN environment variable"
				if !dsnEnvExists {
					if portEnvExists {
						err := resolvedConfig.setPortStr(
							portEnv, "EDGEDB_PORT environment variable",
						)
						if err != nil {
							return nil, err
						}
					}
					if hostEnvExists {
						err := validateHost(host)
						if err != nil {
							return nil, err
						}
						dsn = "edgedb://" + host
						dsnSource = "EDGEDB_HOST environment variable"
					} else {
						dsn = "edgedb://"
						dsnSource = "EDGEDB_PORT environment variable"
					}
				}

				err := parseDSNIntoConfig(&resolvedConfig, dsn, dsnSource)
				if err != nil {
					return nil, err
				}
			} else {
				err := loadCredentialsIntoConfig(
					&resolvedConfig,
					instanceName,
					instanceEnvExists,
					"dsn (parsed as instance name)",
					credentialsFile,
					credFileEnvExists,
					"CredentialsFile option",
				)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if compoundOptionsCount == 0 {
		dir, err := os.Getwd()
		if err != nil {
			return nil, &clientConnectionError{err: err}
		}

		tomlPath := filepath.Join(dir, "edgedb.toml")
		if _, e := os.Stat(tomlPath); os.IsNotExist(e) {
			return nil, &clientConnectionError{
				msg: "no `edgedb.toml` found " +
					"and no connection options specified" +
					" either via arguments to connect API " +
					"or via environment variables " +
					"EDGEDB_HOST/EDGEDB_PORT, EDGEDB_INSTANCE, " +
					"EDGEDB_DSN or EDGEDB_CREDENTIALS_FILE",
			}
		}

		stashDir, err := stashPath(dir)
		if err != nil {
			return nil, &clientConnectionError{err: err}
		}

		if _, e := os.Stat(stashDir); os.IsNotExist(e) {
			return nil, &clientConnectionError{
				msg: "Found `edgedb.toml` " +
					"but the project is not initialized. " +
					"Run `edgedb project init`.",
			}
		}

		data, err := ioutil.ReadFile(filepath.Join(stashDir, "instance-name"))
		if err != nil {
			return nil, &clientConnectionError{err: err}
		}

		instanceName = strings.TrimSpace(string(data))

		err = loadCredentialsIntoConfig(
			&resolvedConfig,
			instanceName,
			true,
			"project linked instance ('"+instanceName+"')",
			"",
			false,
			"",
		)
		if err != nil {
			return nil, err
		}
	}

	host := resolvedConfig.host
	if host == "" {
		host = "localhost"
	}
	port := resolvedConfig.port
	if port == 0 {
		port = 5656
	}

	addrs := []*dialArgs{{
		"tcp",
		fmt.Sprintf("%v:%v", host, port),
	}}

	database := resolvedConfig.database
	if database == "" {
		database = "edgedb"
	}
	user := resolvedConfig.user
	if user == "" {
		user = "edgedb"
	}

	var roots *x509.CertPool
	certData := &resolvedConfig.tlsCAData
	if len(*certData) != 0 {
		roots = x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(*certData)
		if !ok {
			return nil, &configurationError{msg: "invalid certificate data"}
		}
	} else {
		var err error
		roots, err = getSystemCertPool()
		if err != nil {
			return nil, &configurationError{err: err}
		}
	}

	verifyHostname := resolvedConfig.tlsVerifyHostname
	if !resolvedConfig.tlsVerifyHostnameSet {
		verifyHostname = len(resolvedConfig.tlsCAData) == 0
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
				DNSName:       cs.ServerName,
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

	cfg := &connConfig{
		addrs:              addrs,
		user:               user,
		password:           resolvedConfig.password,
		database:           database,
		connectTimeout:     opts.ConnectTimeout,
		waitUntilAvailable: waitUntilAvailable,
		serverSettings:     resolvedConfig.serverSettings,
		tlsConfig:          tlsConfig,
	}

	return cfg, nil
}

func loadCredentialsIntoConfig(
	resolvedConfig *resolvedParams,
	instanceName string,
	instanceIsSet bool,
	instanceNameSource string,
	credentialsFile string,
	credentialsFileIsSet bool,
	credentialsFileSource string,
) error {
	if instanceIsSet && credentialsFileIsSet {
		return &configurationError{msg: "cannot have both instance name " +
			"and credentials file"}
	}

	source := credentialsFileSource
	if instanceIsSet {
		if !isIdentifier.MatchString(instanceName) {
			return &configurationError{
				msg: "invalid instance name '" + instanceName + "'",
			}
		}
		var err error
		credentialsFile, err = findConfigPath(
			"credentials", instanceName+".json",
		)
		if err != nil {
			return &configurationError{msg: err.Error()}
		}
		source = instanceNameSource
	}

	creds, err := readCredentials(credentialsFile)
	if err != nil {
		if credentialsFileIsSet {
			return &configurationError{
				msg: fmt.Sprintf("cannot read credentials of file '%q': %v",
					credentialsFile, err,
				),
			}
		}
		return &configurationError{
			msg: fmt.Sprintf("cannot read credentials of instance %q: %v",
				instanceName, err,
			),
		}
	}

	if err := resolvedConfig.setHost(creds.host, source); err != nil {
		return err
	}
	if err := resolvedConfig.setPort(creds.port, source); err != nil {
		return err
	}
	if err := resolvedConfig.setDatabase(creds.database, source); err != nil {
		return err
	}
	if err := resolvedConfig.setUser(creds.user, source); err != nil {
		return err
	}
	resolvedConfig.setPassword(creds.password, source)
	resolvedConfig.setTLSCAData(creds.certData, source)
	if verifyHostname, verifyHostnameSet :=
		creds.verifyHostname.Get(); verifyHostnameSet {
		resolvedConfig.setTLSVerifyHostname(verifyHostname, source)
	}

	return nil
}

func parseDSNIntoConfig(
	resolvedConfig *resolvedParams,
	dsn string,
	source string,
) error {
	parsed, err := url.Parse(dsn)

	if err != nil {
		return &configurationError{msg: fmt.Sprintf(
			"could not parse DSN %q: %v", dsn, err)}
	}

	if parsed.Scheme != "edgedb" {
		return &configurationError{msg: fmt.Sprintf(
			`invalid DSN: scheme is expected to be "edgedb", got %q`,
			parsed.Scheme,
		)}
	}

	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return &configurationError{msg: fmt.Sprintf(
			"could not parse DSN query parameters %q: %v",
			parsed.RawQuery, err,
		)}
	}
	for k, v := range query {
		if len(v) > 1 {
			return &configurationError{msg: fmt.Sprintf(
				`invalid DSN: duplicate query parameter %q`, k,
			)}
		}
	}

	if host, source, isSet, err := getDSNPart(
		&query, "host", parsed.Hostname(), resolvedConfig.host != "",
	); err != nil || isSet {
		if err != nil {
			return err
		}
		err := resolvedConfig.setHost(host, source)
		if err != nil {
			return err
		}
	}

	if port, source, isSet, err := getDSNPart(
		&query, "port", parsed.Port(), resolvedConfig.port != 0,
	); err != nil || isSet {
		if err != nil {
			return err
		}
		err := resolvedConfig.setPortStr(port, source)
		if err != nil {
			return err
		}
	}

	if database, source, isSet, err := getDSNPart(
		&query, "database",
		strings.TrimPrefix(parsed.Path, "/"),
		resolvedConfig.database != "",
	); err != nil || isSet {
		if err != nil {
			return err
		}
		err := resolvedConfig.setDatabase(
			strings.TrimPrefix(database, "/"), source,
		)
		if err != nil {
			return err
		}
	}

	if user, source, isSet, err := getDSNPart(
		&query, "user", parsed.User.Username(), resolvedConfig.user != "",
	); err != nil || isSet {
		if err != nil {
			return err
		}
		err := resolvedConfig.setUser(user, source)
		if err != nil {
			return err
		}
	}

	parsedPassword, _ := parsed.User.Password()
	if password, source, isSet, err := getDSNPart(
		&query, "password", parsedPassword, resolvedConfig.passwordSet,
	); err != nil || isSet {
		if err != nil {
			return err
		}
		resolvedConfig.setPassword(password, source)
	}

	if certFile, source, isSet, err := getDSNPart(
		&query, "tls_cert_file", "", len(resolvedConfig.tlsCAData) != 0,
	); err != nil || isSet {
		if err != nil {
			return err
		}
		err := resolvedConfig.setTLSCAFile(certFile, source)
		if err != nil {
			return err
		}
	}

	if verifyHostname, source, isSet, err := getDSNPart(
		&query, "tls_verify_hostname", "", resolvedConfig.tlsVerifyHostnameSet,
	); err != nil || isSet {
		if err != nil {
			return err
		}
		err := resolvedConfig.setTLSVerifyHostnameStr(verifyHostname, source)
		if err != nil {
			return err
		}
	}

	serverSettings := make(map[string]string, len(query))
	for k, v := range query {
		serverSettings[k] = v[0]
	}
	resolvedConfig.addServerSettings(serverSettings)

	return nil
}

func getDSNPart(
	query *url.Values,
	paramName string,
	value string,
	isResolved bool,
) (string, string, bool, error) {
	duplicateCount := 0
	if value != "" {
		duplicateCount++
	}
	queryVal, queryValIsSet := query.Get(paramName), query.Has(paramName)
	if queryValIsSet {
		duplicateCount++
		query.Del(paramName)
	}
	queryEnvVal, queryEnvValIsSet :=
		query.Get(paramName+"_env"), query.Has(paramName+"_env")
	if queryEnvValIsSet {
		duplicateCount++
		query.Del(paramName + "_env")
	}
	queryFileVal, queryFileValIsSet :=
		query.Get(paramName+"_file"), query.Has(paramName+"_file")
	if queryFileValIsSet {
		duplicateCount++
		query.Del(paramName + "_file")
	}

	if duplicateCount > 1 {
		var dsnValMsg string
		if value != "" {
			dsnValMsg = fmt.Sprintf(`%v, `, paramName)
		}
		return "", "", false, &configurationError{
			msg: fmt.Sprintf(`invalid DSN: more than one of %v?%v=, `+
				`?%v_env= or ?%v_file= was specified`, dsnValMsg, paramName,
				paramName, paramName),
		}
	}

	param := value
	source := ""
	isSet := false

	if !isResolved {
		if param != "" {
			isSet = true
		}
		if !isSet && queryValIsSet {
			param = queryVal
			source = fmt.Sprintf(` (?%v=)`, paramName)
			isSet = true
		}
		if !isSet && queryEnvValIsSet {
			var envExists bool
			param, envExists = os.LookupEnv(queryEnvVal)
			if !envExists {
				return "", "", false, &configurationError{
					msg: fmt.Sprintf(`'%v_env' environment variable %q `+
						`doesn't exist`, paramName, queryEnvVal),
				}
			}
			source = fmt.Sprintf(` (%v_env: %q)`, paramName, queryEnvVal)
			isSet = true
		}
		if !isSet && queryFileValIsSet {
			data, err := ioutil.ReadFile(queryFileVal)
			if err != nil {
				return "", "", false, &configurationError{
					msg: fmt.Sprintf(`failed to read '%v_file' file %q: %v`,
						paramName, queryFileVal, err),
				}
			}
			param = string(data)
			source = fmt.Sprintf(` (%v_file: %q)`, paramName, queryFileVal)
			isSet = true
		}
	}

	return param, source, isSet, nil
}
