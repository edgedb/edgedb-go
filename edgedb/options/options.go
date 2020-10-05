package options

import (
	"net/url"
	"strconv"
	"strings"
)

// Options for connecting to an EdgeDB server
type Options struct {
	Host     string
	Port     int
	User     string
	Database string
	Password string
	// todo support authentication etc.
}

// FromDSN parses a URI string into an Options struct
func FromDSN(dsn string) Options {
	parsed, err := url.Parse(dsn)
	if err != nil {
		panic(err)
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		panic(err)
	}

	host := strings.Split(parsed.Host, ":")[0]
	db := strings.TrimLeft(parsed.Path, "/")
	password, _ := parsed.User.Password()

	return Options{
		Host:     host,
		Port:     port,
		User:     parsed.User.Username(),
		Database: db,
		Password: password,
	}
}
