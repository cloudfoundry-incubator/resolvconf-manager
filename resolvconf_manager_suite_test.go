package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"
)

func TestResolvconfManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ResolvconfManager Suite")
}

var _ = BeforeSuite(func() {
	err := aptUpdate()
	Expect(err).NotTo(HaveOccurred())
})

func aptUpdate() error {
	cmd := exec.Command("apt", "update")
	ses, err := gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
	Eventually(ses, 5*time.Minute).Should(gexec.Exit(0))
	if err != nil {
		return err
	}
	return nil
}

func installResolvConf() error {
	err := installPackage("resolvconf")
	Expect(err).NotTo(HaveOccurred())

	err = setupStubResolvConfDirectory()
	if err != nil {
		return err
	}

	err = zeroOutOriginalInterface()
	if err != nil {
		return err
	}

	err = bindMount(stubResolvConfDirectory, "/etc")
	if err != nil {
		return err
	}
	cmd := exec.Command("resolvconf", "--enable-updates")
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func installPackage(pkg string) error {
	// When installing resolvconf or openresolv, we expect to see an error trying
	// to create a symlink to /etc/resolv.conf. The error is usually "Device or
	// resource busy".

	cmd := exec.Command("apt", "install", "-y", pkg)
	ses, err := gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
	if err != nil && !isKnownError(string(ses.Out.Contents())) {
		return err
	}

	Eventually(ses, 1*time.Minute).Should(gexec.Exit())
	return nil
}

func purgePackage(pkg string) error {
	cmd := exec.Command("apt", "purge", "-y", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil && !isKnownError(string(output)) {
		return err
	}
	return nil
}

func isKnownError(output string) bool {
	return strings.Contains(output, UnableLocatePackage) ||
		strings.Contains(output, DeviceResourceBusy)
}

func zeroOutOriginalInterface() error {
	fp := filepath.Join(
		"/run",
		"resolvconf",
		"interface",
		"original.resolvconf",
	)
	err := ioutil.WriteFile(
		fp,
		nil,
		0644,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%s'", fp)
	}

	return nil
}

func setupStubResolvConfDirectory() error {
	err := os.MkdirAll(filepath.Join(stubResolvConfDirectory, "resolvconf", "resolv.conf.d"), 0700)
	if err != nil {
		return err
	}

	err = copyFolder(
		filepath.Join("/etc", "resolvconf", "update.d"),
		filepath.Join(stubResolvConfDirectory, "resolvconf", "update.d"),
	)
	if err != nil {
		return err
	}

	for _, f := range []string{"head", "base", "tail"} {
		fp := filepath.Join(
			stubResolvConfDirectory,
			"resolvconf",
			"resolv.conf.d",
			f,
		)
		err = ioutil.WriteFile(
			fp,
			nil,
			0644,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to write file '%s'", fp)
		}
	}

	err = symlink("/run/resolvconf/resolv.conf",
		filepath.Join(stubResolvConfDirectory, "resolv.conf"))
	if err != nil {
		return err
	}

	return nil
}

func copyFolder(src, dest string) error {
	cmd := exec.Command("cp", "-r", src, dest)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func bindMount(src, target string) error {
	cmd := exec.Command("mount", "-o", "bind", src, target)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func umount(target string) error {
	cmd := exec.Command("umount", target)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func symlink(src, target string) error {
	cmd := exec.Command("ln", "-sf", src, target)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func serverIsFirst(address, contents string) bool {
	allNameservers := parseNameservers(contents)

	if len(allNameservers) == 0 {
		return false
	}

	return allNameservers[0] == address
}

func baseMatches(addresses []string, contents string) bool {
	allNameservers := parseNameservers(contents)
	configuredNameservers := make(map[string]bool)

	for _, address := range addresses {
		configuredNameservers[address] = false
	}

	for _, address := range allNameservers {
		configuredNameservers[address] = true
	}

	for _, value := range configuredNameservers {
		if value == false {
			return false
		}
	}

	return true
}

func parseNameservers(contents string) []string {
	nameserverLineRegex := regexp.MustCompile("^nameserver (.+)")
	lines := strings.Split(contents, "\n")

	nameservers := []string{}
	for _, l := range lines {
		server := nameserverLineRegex.FindStringSubmatch(l)
		if len(server) != 2 {
			continue
		}
		nameservers = append(nameservers, server[1])
	}

	return nameservers
}
