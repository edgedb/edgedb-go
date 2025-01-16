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
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/geldata/gel-go/internal/geltypes"
	"github.com/geldata/gel-go/internal/snc"
	"github.com/sigurn/crc16"
)

var (
	isDSNLike      = regexp.MustCompile(`(?i)^[a-z]+://`)
	instanceNameRe = regexp.MustCompile(
		`^(\w(?:-?\w)*)$`,
	)
	cloudInstanceNameRe = regexp.MustCompile(
		`^([A-Za-z0-9_\-](?:-?[A-Za-z_0-9\-])*)/` +
			`([A-Za-z0-9](?:-?[A-Za-z0-9])*)$`,
	)
	domainLabelMaxLength              = 63
	crcTable             *crc16.Table = crc16.MakeTable(crc16.CRC16_XMODEM)
	base64Encoding                    = base64.URLEncoding.WithPadding(
		base64.NoPadding,
	)
)

type connConfig struct {
	addr               dialArgs
	user               string
	password           string
	database           string
	branch             string
	connectTimeout     time.Duration
	waitUntilAvailable time.Duration
	tlsCAData          []byte
	tlsSecurity        string
	tlsServerName      string
	serverSettings     *snc.ServerSettings
	secretKey          string
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
		ServerName: c.tlsServerName,
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
	host               cfgVal // string
	port               cfgVal // int
	database           cfgVal // string
	user               cfgVal // string
	password           cfgVal // OptionalStr
	tlsCAData          cfgVal // []byte
	tlsSecurity        cfgVal // string
	tlsServerName      cfgVal // string
	waitUntilAvailable cfgVal // time.Duration
	serverSettings     *snc.ServerSettings
	secretKey          cfgVal // string
	profile            cfgVal // string
	instance           cfgVal // string
	org                cfgVal // string
}

func (r *configResolver) setInstance(val, source string) error {
	if r.instance.val != nil {
		return nil
	}

	match := instanceNameRe.FindStringSubmatch(val)
	if len(match) == 0 {
		match = cloudInstanceNameRe.FindStringSubmatch(val)
		if len(match) == 0 || strings.Contains(match[1], "--") {
			return fmt.Errorf("invalid instance name %q", val)
		}
		r.org = cfgVal{val: match[1], source: source}
		r.instance = cfgVal{val: match[2], source: source}
	} else {
		r.instance = cfgVal{val: match[1], source: source}
	}

	return nil
}

func (r *configResolver) setProfile(val, source string) {
	if r.profile.val != nil {
		return
	}

	r.profile = cfgVal{val: val, source: source}
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
	data, err := os.ReadFile(file)
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

func (r *configResolver) setTLSServerName(val string, source string) error {
	if r.tlsServerName.val != nil {
		return nil
	}

	r.tlsServerName = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) setWaitUntilAvailable(
	val time.Duration,
	source string,
) error {
	if r.waitUntilAvailable.val != nil {
		return nil
	}

	r.waitUntilAvailable = cfgVal{val: val, source: source}
	return nil
}

func (r *configResolver) setWaitUntilAvailableStr(val, source string) error {
	d, err := geltypes.ParseDuration(val)
	if err != nil {
		return fmt.Errorf("invalid WaitUntilAvailable: %w", err)
	}

	return r.setWaitUntilAvailable(time.Duration(1_000*d), source)
}

func (r *configResolver) setSecretKey(val, source string) error {
	if r.secretKey.val != nil {
		return nil
	}

	r.secretKey = cfgVal{val: val, source: source}
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

func (r *configResolver) resolveOptions(
	opts *Options,
	paths *cfgPaths,
) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid gel.Options: %w", err)
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

	if opts.Branch != "" {
		if e := r.setDatabase(opts.Branch, "Branch options"); e != nil {
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

	if opts.WaitUntilAvailable != 0 {
		e := r.setWaitUntilAvailable(
			opts.WaitUntilAvailable,
			"WaitUntilAvailable Options",
		)
		if e != nil {
			return e
		}
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

	if opts.TLSOptions.ServerName != "" {
		secSources = append(secSources, "TLSOptions.ServerName")
		err = r.setTLSServerName(
			opts.TLSOptions.ServerName,
			"TLSOptions.ServerName options",
		)
	}

	if len(secSources) > 1 {
		return fmt.Errorf(
			"mutually exclusive options set in Options: %v",
			englishList(secSources, "and"))
	}

	if opts.SecretKey != "" {
		err = r.setSecretKey(opts.SecretKey, "SecretKey option")
		if err != nil {
			return err
		}
	}

	if r.secretKey.val != nil && r.instance.val != nil {
		if r.org.val != nil {
			err := r.parseCloudInstanceNameIntoConfig(
				"SecretKey option",
				paths,
			)
			if err != nil {
				return err
			}
		}
	}

	r.addServerSettings(opts.ServerSettings)
	return nil
}

func queryContains(k string, m map[string]string) bool {
	return inMap(k, m) ||
		inMap(fmt.Sprintf("%s_env", k), m) ||
		inMap(fmt.Sprintf("%s_file", k), m)
}

func (r *configResolver) resolveDSN(
	dsn, source string,
	paths *cfgPaths,
) (err error) {
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
	if queryContains("branch", query) {
		if queryContains("database", query) || db != "" {
			return fmt.Errorf(
				"`database` and `branch` " +
					"cannot be present at the same time")
		}

		val, err = popDSNValue(query, db, "branch", r.database.val == nil)
		if err != nil {
			return err
		} else if val.val != nil {
			br := strings.TrimPrefix(val.val.(string), "/")
			if e := r.setDatabase(br, source+val.source); e != nil {
				return e
			}
		}
	} else {
		val, err = popDSNValue(
			query, db, "database", r.database.val == nil)
		if err != nil {
			return err
		} else if val.val != nil {
			db := strings.TrimPrefix(val.val.(string), "/")
			if e := r.setDatabase(db, source+val.source); e != nil {
				return e
			}
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
		if paths.testDir != "" {
			val.val = filepath.Join(paths.testDir, val.val.(string))
		}
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

	val, err = popDSNValue(
		query,
		"",
		"tls_server_name",
		r.tlsServerName.val == nil,
	)
	if err != nil {
		return err
	}
	if val.val != nil {
		err = r.setTLSServerName(val.val.(string), source+val.source)
		if err != nil {
			return err
		}
	}

	val, err = popDSNValue(
		query,
		"",
		"wait_until_available",
		r.waitUntilAvailable.val == nil,
	)
	if err != nil {
		return err
	}
	if val.val != nil {
		err = r.setWaitUntilAvailableStr(val.val.(string), source+val.source)
		if err != nil {
			return err
		}
	}

	val, err = popDSNValue(query, "", "secret_key", r.secretKey.val == nil)
	if err != nil {
		return err
	}
	if val.val != nil {
		err = r.setSecretKey(val.val.(string), source+val.source)
		if err != nil {
			return err
		}
	}

	r.addServerSettingsStr(query)
	return nil
}

func (r *configResolver) resolveCredentials(
	credentials, source string,
	paths *cfgPaths,
) error {
	if r.instance.val != nil && credentials != "" {
		return errors.New("cannot have both instance name and " +
			"credentials file")
	}

	if r.instance.val != nil {
		if r.org.val != nil {
			return r.parseCloudInstanceNameIntoConfig(source, paths)
		}

		dir, err := paths.CfgDir()
		if err != nil {
			return err
		}
		credentials = filepath.Join(
			dir,
			"credentials",
			r.instance.val.(string)+".json",
		)
	}

	creds, err := readCredentials(credentials)
	if err != nil {
		if r.instance.val != nil {
			return fmt.Errorf(
				"cannot read credentials for instance %q: %w",
				r.instance.val.(string),
				err,
			)
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

	if br, ok := creds.branch.Get(); ok && br != "" {
		if e := r.setDatabase(br, source); e != nil {
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
	db, dbOk := os.LookupEnv("EDGEDB_DATABASE")
	if dbOk {
		err := r.setDatabase(db, "EDGEDB_DATABASE environment variable")
		if err != nil {
			return false, err
		}
	}

	branchEnvVarName, branch, branchOk := lookupGelOrEdgedbEnv("_BRANCH")
	if branchOk {
		if dbOk {
			return false, fmt.Errorf(
				"mutually exclusive options EDGEDB_DATABASE and "+
					"%s environment variables are set",
				branchEnvVarName,
			)
		}
		err := r.setDatabase(
			branch,
			fmt.Sprintf("%s environment variable", branchEnvVarName),
		)
		if err != nil {
			return false, err
		}
	}

	if envvarname, user, ok := lookupGelOrEdgedbEnv("_USER"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		err := r.setUser(user, source)
		if err != nil {
			return false, err
		}
	}

	if envvarname, pwd, ok := lookupGelOrEdgedbEnv("_PASSWORD"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		r.setPassword(pwd, source)
	}

	if name, wua, ok := lookupGelOrEdgedbEnv("_WAIT_UNTIL_AVAILABLE"); ok {
		source := fmt.Sprintf("%s environment variable", name)
		err := r.setWaitUntilAvailableStr(wua, source)
		if err != nil {
			return false, err
		}
	}

	var tlsCaSources []string

	if envvarname, caString, ok := lookupGelOrEdgedbEnv("_TLS_CA"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		r.setTLSCAData([]byte(caString), source)
		tlsCaSources = append(tlsCaSources, envvarname)
	}

	if envvarname, file, ok := lookupGelOrEdgedbEnv("_TLS_CA_FILE"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		e := r.setTLSCAFile(file, source)
		tlsCaSources = append(tlsCaSources, envvarname)
		if e != nil {
			return false, e
		}
	}

	if envvarname, val, ok := lookupGelOrEdgedbEnv("_TLS_SERVER_NAME"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		e := r.setTLSServerName(val, source)
		if e != nil {
			return false, e
		}
	}

	if len(tlsCaSources) > 1 {
		return false, fmt.Errorf(
			"mutually exclusive environment variables set: %v",
			englishList(tlsCaSources, "and"))
	}

	if name, verify, ok := lookupGelOrEdgedbEnv("_CLIENT_TLS_SECURITY"); ok {
		source := fmt.Sprintf("%s environment variable", name)
		if e := r.setTLSSecurity(verify, source); e != nil {
			return false, e
		}
	}

	var names []string
	dsnEnvVarName, dsn, dsnOk := lookupGelOrEdgedbEnv("_DSN")
	if dsnOk {
		names = append(names, dsnEnvVarName)
	}
	instanceEnvVarName, instance, instanceOk :=
		lookupGelOrEdgedbEnv("_INSTANCE")
	if instanceOk {
		names = append(names, instanceEnvVarName)
		source := fmt.Sprintf("%s environment variable", instanceEnvVarName)
		err := r.setInstance(instance, source)
		if err != nil {
			return false, err
		}
	}
	credentialsEnvVarName, credentials, credsOk :=
		lookupGelOrEdgedbEnv("_CREDENTIALS_FILE")
	if credsOk {
		names = append(names, credentialsEnvVarName)
	}

	hostEnvVarName, host, hostOk := lookupGelOrEdgedbEnv("_HOST")
	if hostOk {
		names = append(names, hostEnvVarName)
	}
	portEnvVarName, port, portOk := lookupGelOrEdgedbEnv("_PORT")
	if portOk && strings.HasPrefix(port, "tcp://") {
		// EDGEDB_PORT is set by 'docker --link' so ignore and warn
		log.Printf(
			"Warning: ignoring %s in 'tcp://host:port' format\n",
			portEnvVarName,
		)
		portOk = false
	}

	if !hostOk && portOk {
		names = append(names, portEnvVarName)
	}

	if len(names) > 1 {
		return false, fmt.Errorf(
			"mutually exclusive environment variables set: %v",
			englishList(names, "and"))
	}

	if envvarname, profile, ok := lookupGelOrEdgedbEnv("_CLOUD_PROFILE"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		r.setProfile(profile, source)
	}

	if envvarname, secretKey, ok := lookupGelOrEdgedbEnv("_SECRET_KEY"); ok {
		source := fmt.Sprintf("%s environment variable", envvarname)
		e := r.setSecretKey(secretKey, source)
		if e != nil {
			return false, e
		}
	}

	switch {
	case hostOk || portOk:
		if portOk {
			err := r.setPortStr(
				port,
				fmt.Sprintf("%s environment variable", portEnvVarName),
			)
			if err != nil {
				return false, err
			}
		}

		if hostOk {
			err := r.setHost(
				host,
				fmt.Sprintf("%s environment variable", hostEnvVarName),
			)
			if err != nil {
				return false, err
			}
		}
	case dsnOk:
		e := r.resolveDSN(
			dsn,
			fmt.Sprintf("%s environment variable", dsnEnvVarName),
			paths,
		)
		if e != nil {
			return false, e
		}
	case instanceOk || credsOk:
		source := fmt.Sprintf("%s environment variable", credentialsEnvVarName)
		if instanceOk {
			source = fmt.Sprintf("%s environment variable", instanceEnvVarName)
		}
		err := r.resolveCredentials(
			credentials,
			source,
			paths,
		)
		if err != nil {
			return false, err
		}
	default:
		return false, nil
	}

	return true, nil
}

func (r *configResolver) resolveTOML(paths *cfgPaths) error {
	toml, err := findEdgeDBTOML(paths)
	if err != nil {
		return err
	}

	stashDir, err := stashPath(filepath.Dir(toml), paths)
	if err != nil {
		return err
	}

	if !exists(stashDir) {
		return errors.New("Found `gel.toml` " +
			"but the project is not initialized. Run `gel project init`.")
	}

	instance, err := os.ReadFile(filepath.Join(stashDir, "instance-name"))
	if err != nil {
		return err
	}

	profile, err := os.ReadFile(filepath.Join(stashDir, "cloud-profile"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	r.setProfile(strings.TrimSpace(string(profile)), "project")

	db, err := os.ReadFile(filepath.Join(stashDir, "database"))
	if err == nil {
		database := strings.TrimSpace(string(db))
		if err = r.setDatabase(database, "database"); err != nil {
			return err
		}
	}

	return r.setInstance(
		strings.TrimSpace(string(instance)),
		"project link",
	)
}

func lookupGelOrEdgedbEnv(name string) (string, string, bool) {
	gelName := fmt.Sprintf("GEL%s", name)
	edbName := fmt.Sprintf("EDGEDB%s", name)
	gelVal, gelOk := os.LookupEnv(gelName)
	edbVal, edbOk := os.LookupEnv(edbName)

	if gelOk && edbOk {
		log.Printf(
			"Both %s and %s are set. %s will be ignored.\n",
			gelName, edbName, edbName,
		)
	}

	if gelOk {
		return gelName, gelVal, true
	} else if edbOk {
		return edbName, edbVal, true
	}

	return "", "", false
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
	branch := "__default__"
	if r.database.val != nil {
		database = r.database.val.(string)
		branch = database
	}

	user := "edgedb"
	if r.user.val != nil {
		user = r.user.val.(string)
	}

	waitUntilAvailable := 30 * time.Second
	if r.waitUntilAvailable.val != nil {
		waitUntilAvailable = r.waitUntilAvailable.val.(time.Duration)
	}

	var certData []byte
	if r.tlsCAData.val != nil {
		certData = r.tlsCAData.val.([]byte)
	}

	tlsSecurity := "default"
	if r.tlsSecurity.val != nil {
		tlsSecurity = r.tlsSecurity.val.(string)
	}

	tlsServerName := ""
	if r.tlsServerName.val != nil {
		tlsServerName = r.tlsServerName.val.(string)
	}

	secretKey := ""
	if r.secretKey.val != nil {
		secretKey = r.secretKey.val.(string)
	}

	clientSecurityVarName, security, err := getEnvVarSetting(
		"_CLIENT_SECURITY",
		"default",
		"default",
		"insecure_dev_mode",
		"strict",
	)
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
				"%s=strict but tls_security=%v, "+
					"tls_security must be set to strict "+
					"when %s is strict",
				clientSecurityVarName,
				tlsSecurity,
				clientSecurityVarName,
			)
		}
	}

	if tlsSecurity == "default" {
		if len(certData) == 0 {
			tlsSecurity = "strict"
		} else {
			tlsSecurity = "no_host_verification"
		}
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
		branch:             branch,
		connectTimeout:     opts.ConnectTimeout,
		waitUntilAvailable: waitUntilAvailable,
		serverSettings:     r.serverSettings,
		tlsCAData:          certData,
		tlsSecurity:        tlsSecurity,
		tlsServerName:      tlsServerName,
		secretKey:          secretKey,
	}, nil
}

func getEnvVarSetting(
	name,
	defalt string,
	values ...string,
) (string, string, error) {
	envvarname, value, ok := lookupGelOrEdgedbEnv(name)
	if !ok || value == "default" || value == "" {
		return envvarname, defalt, nil
	}

	for _, v := range values {
		if value == v {
			return envvarname, value, nil
		}
	}

	return "", "", fmt.Errorf(
		"environment variable %v should be one of %v, got: %q",
		envvarname, englishList(append(values, "default"), "or"), value)
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

	if instance != "" {
		err := cfg.setInstance(instance, "dsn (parsed as instance name)")
		if err != nil {
			return nil, err
		}
	}

	var names []string
	if dsn != "" || instance != "" {
		names = append(names, "dsn")
	}
	if opts.Credentials != nil {
		names = append(names, "gel.Options.Credentials")
	}
	if opts.CredentialsFile != "" {
		names = append(names, "gel.Options.CredentialsFile")
	}
	if opts.Host != "" {
		names = append(names, "gel.Options.Host")
	} else if opts.Port != 0 {
		names = append(names, "gel.Options.Port")
	}
	if len(names) > 1 {
		return nil, fmt.Errorf(
			"mutually exclusive connection options specified: %v",
			englishList(names, "and"))
	}

	if e := cfg.resolveOptions(opts, paths); e != nil {
		return nil, e
	}

	switch {
	case opts.Host != "" || opts.Port != 0:
		// stop here since there is a host or port
	case dsn != "":
		if e := cfg.resolveDSN(dsn, "DSN option", paths); e != nil {
			return nil, e
		}
	case instance != "" || opts.CredentialsFile != "":
		source := "CredentialsFile option"
		if instance != "" {
			source = "dsn (parsed as instance name)"
		}
		err := cfg.resolveCredentials(opts.CredentialsFile, source, paths)
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

		err = cfg.resolveTOML(paths)
		if errors.Is(err, errNoTOMLFound) {
			return nil, errors.New(
				"no `gel.toml` found and no connection options " +
					"specified either via arguments to connect API " +
					"or via environment variables " +
					"GEL_HOST/GEL_PORT, GEL_INSTANCE, " +
					"GEL_DSN or GEL_CREDENTIALS_FILE",
			)
		}
		if err != nil {
			return nil, err
		}

		source := fmt.Sprintf(
			"project linked instance (%q)",
			cfg.instance.val.(string),
		)
		err = cfg.resolveCredentials("", source, paths)
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

	if uri.Scheme != "edgedb" && uri.Scheme != "gel" {
		return nil, nil, fmt.Errorf(
			`scheme is expected to be "gel", got %q`, uri.Scheme)
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
	"branch":       {"branch", "branch_env", "branch_file"},
	"user":         {"user", "user_env", "user_file"},
	"password":     {"password", "password_env", "password_file"},
	"tls_ca_file":  {"tls_ca_file", "tls_ca_file_env"},
	"tls_security": {"tls_security", "tls_security_env", "tls_security_file"},
	"tls_server_name": {
		"tls_server_name",
		"tls_server_name_env",
		"tls_server_name_file",
	},
	"tls_verify_hostname": {
		"tls_verify_hostname",
		"tls_verify_hostname_env",
		"tls_verify_hostname_file",
	},
	"wait_until_available": {
		"wait_until_available",
		"wait_until_available_env",
		"wait_until_available_file",
	},
	"secret_key": {"secret_key", "secret_key_env", "secret_key_file"},
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
		data, err := os.ReadFile(val)
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
	testDir   string
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
		tomlPath := filepath.Join(dir, "gel.toml")
		if _, e := os.Stat(tomlPath); os.IsNotExist(e) {
			tomlPath = filepath.Join(dir, "edgedb.toml")
			if _, e := os.Stat(tomlPath); os.IsNotExist(e) {
				parent := filepath.Dir(dir)
				// Stop searching when dir is the root directory.
				if parent == dir {
					return "", errNoTOMLFound
				}

				pDev, err := device(parent)
				if err != nil {
					return "", fmt.Errorf(
						"searching for gel.toml in or above %q: %w",
						filepath.Dir(tomlPath), err)
				}

				// Stop searching at file system boundaries.
				if pDev != dev {
					if err == nil { // nolint:govet
						err = errNoTOMLFound
					}
					return "", fmt.Errorf(
						"%w: stopped searching for gel.toml "+
							"at file system boundary %q", err, dir)
				}

				dir = parent
				dev = pDev
				continue
			}
		}
		return tomlPath, nil
	}
}

func jwtBase64Decode(data []byte) (map[string]interface{}, error) {
	decoded := make([]byte, base64Encoding.DecodedLen(len(data)))
	_, err := base64Encoding.Decode(decoded, data)
	if err != nil {
		return nil, err
	}

	var jwt map[string]interface{}
	err = json.Unmarshal(decoded, &jwt)
	if err != nil {
		return nil, err
	}

	return jwt, nil
}

func (r *configResolver) parseCloudInstanceNameIntoConfig(
	source string,
	paths *cfgPaths,
) (e error) {
	if r.instance.val == nil {
		return fmt.Errorf("missing instance")
	}
	inst, instOk := r.instance.val.(string)
	if !instOk {
		return fmt.Errorf("instance is not a string")
	}
	inst = strings.ToLower(inst)

	if r.org.val == nil {
		return fmt.Errorf("missing org")
	}
	org, orgOk := r.org.val.(string)
	if !orgOk {
		return fmt.Errorf("org is not a string")
	}
	org = strings.ToLower(org)

	label := fmt.Sprintf("%s--%s", inst, org)
	if len(label) > domainLabelMaxLength {
		return fmt.Errorf(
			"invalid instance name: cloud instance name length"+
				" cannot exceed %d characters: %s/%s",
			domainLabelMaxLength-1, org, inst,
		)
	}

	var secretKey string
	if r.secretKey.val != nil {
		secretKey = r.secretKey.val.(string)
	} else {
		errMsg := "Cannot connect to cloud instances without secret key: %w"

		dir, err := paths.CfgDir()
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		profile := "default"
		if r.profile.val != nil {
			profile = r.profile.val.(string)
		} else {
			if name, p, ok := lookupGelOrEdgedbEnv("_CLOUD_PROFILE"); ok {
				r.setProfile(p, fmt.Sprintf("%s environment variable", name))
				profile = r.profile.val.(string)
			}
		}

		path := path.Join(dir, "cloud-credentials", profile+".json")
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}
		defer func() {
			fErr := f.Close()
			if e == nil {
				e = fErr
			}
		}()

		data, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		var creds map[string]interface{}
		err = json.Unmarshal(data, &creds)
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}

		key, ok := creds["secret_key"]
		if !ok {
			return fmt.Errorf(errMsg, fmt.Errorf(
				"access_token not found in profile "+
					"%q's credentials file %q",
				profile, path))
		}

		secretKey, ok = key.(string)
		if !ok {
			return fmt.Errorf(errMsg, fmt.Errorf(
				"access_token in profile %q's credential file %q "+
					"is the wrong type, expected string but got %T",
				profile, path, key))
		}

		err = r.setSecretKey(secretKey, "cloud-credentials/"+profile+".json")
		if err != nil {
			return fmt.Errorf(errMsg, err)
		}
	}

	data := strings.Split(secretKey, ".")
	if len(data) < 2 {
		return fmt.Errorf("Invalid secret key: JWT is missing parts")
	}

	jwt, err := jwtBase64Decode([]byte(data[1]))
	if err != nil {
		return fmt.Errorf("Invalid secret key: %w", err)
	}

	iss, ok := jwt["iss"]
	if !ok {
		return fmt.Errorf("Invalid secret key: iss is missing")
	}

	dnsZone, ok := iss.(string)
	if !ok {
		return fmt.Errorf(
			"Invalid secret key: iss is the wrong type, "+
				"expected string but got %T",
			iss)
	}

	crc := crc16.Checksum([]byte(fmt.Sprintf("%s/%s", org, inst)), crcTable)
	host := fmt.Sprintf("%s.c-%02d.i.%s", label, crc%100, dnsZone)
	return r.setHost(host, source)
}
