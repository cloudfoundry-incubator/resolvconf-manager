package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/pkg/errors"
)

var nameserverLineRegex = regexp.MustCompile("^nameserver (.+)")

const ResolvConfHeadFile = "/etc/resolvconf/resolv.conf.d"
const OpenResolvConfFile = "/etc/resolvconf.conf"

const address = "99.99.99.99"

func main() {
	fmt.Printf(`
IsResolvconf: %v
IsOpenresolv: %v
`, IsResolvconf(), IsOpenresolv())

	switch {
	case IsResolvconf():
		err := WriteResolvConfHead()
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfHeadFile, err)
		}
	case IsOpenresolv():
		err := WriteOpenResolvConf()
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfHeadFile, err)
		}
	}

	// call `resolvconf -u`
}

func IsResolvconf() bool {
	is, err := exists(ResolvConfHeadFile)
	if err != nil {
		log.Fatalf("Error occurred while checking for '%s': %v", ResolvConfHeadFile, err)
	}
	return is
}

func IsOpenresolv() bool {
	is, err := exists(OpenResolvConfFile)
	if err != nil {
		log.Fatalf("Error occurred while checking for '%s': %v", OpenResolvConfFile, err)
	}
	return is
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func WriteResolvConfHead() error {
	contents := fmt.Sprintf(
		`# This file was automatically updated by bosh-dns

nameserver %s
`, address)

	return writeFile(ResolvConfHeadFile, contents)
}

func WriteOpenResolvConf() error {
	contents := fmt.Sprintf(
		`resolv_conf=/etc/resolv.conf
name_servers=%s`, address)

	return writeFile(OpenResolvConfFile, contents)
}

func writeFile(filename, contents string) error {
	err := ioutil.WriteFile(filename, []byte(contents), 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%s'", filename)
	}

	return nil
}
