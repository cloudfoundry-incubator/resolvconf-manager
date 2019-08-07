package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/cloudfoundry-incubator/resolvconf-manager"
)

const DeviceResourceBusy = "Device or resource busy"
const UnableLocatePackage = "Unable to locate package"

var _ = Describe("Main", func() {
	BeforeEach(func() {
		dat, err := ioutil.ReadFile("/etc/resolv.conf")
		Expect(err).NotTo(HaveOccurred())
		fmt.Println(string(dat))
	})

	Context("IsResolvconf", func() {
		BeforeEach(func() {
			err := InstallResolvConf()
			Expect(err).NotTo(HaveOccurred())
			err = PurgePackage("openresolv")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := Umount("/etc")
			Expect(err).NotTo(HaveOccurred())
		})

		It("detects the resolvconf package is installed", func() {
			isResolvconf, err := IsResolvconf()
			Expect(err).NotTo(HaveOccurred())
			Expect(isResolvconf).To(BeTrue())

			isOpenresolv, err := IsOpenresolv()
			Expect(err).NotTo(HaveOccurred())
			Expect(isOpenresolv).To(BeFalse())
		})

		It("updates bosh-dns", func() {
			cli, err := gexec.Build("github.com/cloudfoundry-incubator/resolvconf-manager")
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(cli, "99.99.99.99")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			dat, err := ioutil.ReadFile("/etc/resolv.conf")
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				dat, err = ioutil.ReadFile("/etc/resolv.conf")
				Expect(err).NotTo(HaveOccurred())
				return serverIsFirst("99.99.99.99", string(dat))
			}, 10*time.Second).Should(BeTrue())

			cmd = exec.Command(cli, "98.98.98.98")
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Eventually(func() bool {
				dat, err = ioutil.ReadFile("/etc/resolv.conf")
				Expect(err).NotTo(HaveOccurred())
				return serverIsFirst("98.98.98.98", string(dat))
			}, 10*time.Second).Should(BeTrue())
		})

		It("errors if no input is provided", func() {
			cli, err := gexec.Build("github.com/cloudfoundry-incubator/resolvconf-manager")
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(cli)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session).To(gbytes.Say(
				`resolvconf-manager \[address\]

Update resolv.conf to have the address provided as the first entry in /etc/resolv.conf
`))
		})
	})

	Context("IsOpenresolv", func() {
		BeforeEach(func() {
			err := InstallPackage("openresolv")
			Expect(err).NotTo(HaveOccurred())
			err = PurgePackage("resolvconf")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
		})

		It("detects the openresolv package is installed", func() {
			isResolvconf, err := IsResolvconf()
			Expect(err).NotTo(HaveOccurred())
			Expect(isResolvconf).To(BeFalse())

			isOpenresolv, err := IsOpenresolv()
			Expect(err).NotTo(HaveOccurred())
			Expect(isOpenresolv).To(BeTrue())
		})

		// Do stuff
	})
})

func PurgePackage(pkg string) error {
	fmt.Printf("Purging %s...\n", pkg)
	cmd := exec.Command("apt", "purge", "-y", pkg)
	output, err := cmd.CombinedOutput()
	if err != nil && !IsKnownError(string(output)) {
		return err
	}
	return nil
}

func IsKnownError(output string) bool {
	return strings.Contains(output, UnableLocatePackage) ||
		strings.Contains(output, DeviceResourceBusy)
}

const stubResolvConfDirectory = "/dir"

func InstallResolvConf() error {
	err := InstallPackage("resolvconf")
	Expect(err).NotTo(HaveOccurred())

	err = SetupStubResolvConfDirectory()
	if err != nil {
		return err
	}
	err = BindMount(filepath.Join(stubResolvConfDirectory, "resolvconf"), "/etc/resolvconf")
	if err != nil {
		return err
	}
	err = Umount("/etc/resolvconf")
	if err != nil {
		return err
	}
	err = BindMount(stubResolvConfDirectory, "/etc")
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

func CopyFolder(src, dest string) error {
	cmd := exec.Command("cp", "-r", src, dest)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func SetupStubResolvConfDirectory() error {
	err := os.MkdirAll(filepath.Join(stubResolvConfDirectory), 0700)
	if err != nil {
		return err
	}

	err = CopyFolder("/etc/resolvconf", filepath.Join(stubResolvConfDirectory, "resolvconf"))
	if err != nil {
		return err
	}
	err = Symlink("/run/resolvconf/resolv.conf", filepath.Join(stubResolvConfDirectory, "resolv.conf"))
	if err != nil {
		return err
	}
	return nil
}

func InstallPackage(pkg string) error {
	// When installing resolvconf or openresolv, we expect to see an error trying
	// to create a symlink to /etc/resolv.conf. The error is usually "Device or
	// resource busy".

	fmt.Printf("Installing %s...", pkg)
	cmd := exec.Command("apt", "install", "-y", pkg)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil && !IsKnownError(string(ses.Out.Contents())) {
		return err
	}

	Eventually(ses, 1*time.Minute).Should(gexec.Exit())
	return nil
}

func BindMount(src, target string) error {
	cmd := exec.Command("mount", "-o", "bind", src, target)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func Symlink(src, target string) error {
	cmd := exec.Command("ln", "-sf", src, target)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func Umount(target string) error {
	cmd := exec.Command("umount", target)
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	if err != nil {
		return err
	}

	Eventually(ses).Should(gexec.Exit(0))
	return nil
}

func serverIsFirst(address, contents string) bool {
	var nameserverLineRegex = regexp.MustCompile("^nameserver (.+)")
	lines := strings.Split(contents, "\n")

	for _, l := range lines {
		if !nameserverLineRegex.MatchString(l) {
			continue
		}

		if l == fmt.Sprintf("nameserver %s", address) {
			return true
		}

		return false
	}

	return false
}
