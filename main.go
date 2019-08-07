package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/pkg/errors"
)

var nameserverLineRegex = regexp.MustCompile("^nameserver (.+)")

const ResolvConfHeadFile = "/etc/resolvconf/resolv.conf.d/head"
const OpenResolvConfFile = "/etc/resolvconf.conf"

func main() {
	if len(os.Args) != 2 {
		fmt.Print(`resolvconf-manager [address]

Update resolv.conf to have the address provided as the first entry in /etc/resolv.conf
`)
		os.Exit(1)
	}

	address := os.Args[1]

	isResolvconf, err := IsResolvconf()
	if err != nil {
		log.Fatalf("Error occurred while checking for '%s': %s", ResolvConfHeadFile, err)
	}

	isOpenresolv, err := IsOpenresolv()
	if err != nil {
		log.Fatalf("Error occurred while checking for '%s': %s", OpenResolvConfFile, err)
	}

	fmt.Printf(`
IsResolvconf: %v
IsOpenresolv: %v
`, isResolvconf, isOpenresolv)

	switch {
	case isResolvconf:
		err := WriteResolvConfHead(address)
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfHeadFile, err)
		}
	case isOpenresolv:
		err := WriteOpenResolvConf(address)
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfHeadFile, err)
		}
	}

	// call `resolvconf -u`
	cmd := exec.Command("resolvconf", "-u")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while runnning resolvconf -u '%s': %s", err, output)
	}
}

func IsResolvconf() (bool, error) {
	return exists(ResolvConfHeadFile)
}

func IsOpenresolv() (bool, error) {
	return exists(OpenResolvConfFile)
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

func WriteResolvConfHead(address string) error {
	contents := fmt.Sprintf(
		`# This file was automatically updated by bosh-dns

nameserver %s
`, address)

	return writeFile(ResolvConfHeadFile, contents)
}

func WriteOpenResolvConf(address string) error {
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
