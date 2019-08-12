package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var nameserverLineRegex = regexp.MustCompile("^nameserver (.+)")

const ResolvConfHeadFile = "/etc/resolvconf/resolv.conf.d/head"
const ResolvConfBaseFile = "/etc/resolvconf/resolv.conf.d/base"
const OpenResolvConfFile = "/etc/resolvconf.conf"

func main() {
	head, base, err := parseArgs()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	headAddress := *head
	baseAddresses := *base

	isResolvconf, err := IsResolvconf()
	if err != nil {
		log.Fatalf("Error occurred while checking for '%s': %s", ResolvConfHeadFile, err)
	}

	isOpenresolv, err := IsOpenresolv()
	if err != nil {
		log.Fatalf("Error occurred while checking for '%s': %s", OpenResolvConfFile, err)
	}

	switch {
	case isResolvconf:
		err := WriteResolvConfHead(headAddress)
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfHeadFile, err)
		}
		err = WriteResolvConfBase(baseAddresses)
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfBaseFile, err)
		}
	case isOpenresolv:
		err := WriteOpenResolvConf(headAddress)
		if err != nil {
			log.Fatalf("Error occurred while writing '%s': %v", ResolvConfHeadFile, err)
		}
	}

	cmd := exec.Command("resolvconf", "-u")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error while runnning resolvconf -u '%s': %s", err, output)
	}
}

func parseArgs() (*string, *string, error) {
	head := flag.String("head", "", "ip address to prepend as first nameserver")
	base := flag.String("base", "", "ip address to append as a nameserver")
	flag.Parse()

	if *head == "" && *base == "" {
		return nil, nil, errors.New("either 'head' or 'base' is required")
	}

	return head, base, nil
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
	if address == "" {
		return nil
	}

	contents := fmt.Sprintf(
		`# This file was automatically updated by bosh-dns

nameserver %s
`, address)

	return writeFile(ResolvConfHeadFile, contents)
}

func WriteResolvConfBase(addresses string) error {
	if addresses == "" {
		return nil
	}

	as := strings.Split(addresses, ",")
	nas := []string{}
	for _, a := range as {
		nas = append(nas, fmt.Sprintf("nameserver %s", a))
	}
	contents := strings.Join(nas, "\n")

	return writeFile(ResolvConfBaseFile, contents)
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
