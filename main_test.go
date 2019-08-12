package main_test

import (
	"io/ioutil"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/cloudfoundry-incubator/resolvconf-manager"
)

const DeviceResourceBusy = "Device or resource busy"
const UnableLocatePackage = "Unable to locate package"
const stubResolvConfDirectory = "/dir"

var _ = Describe("Main", func() {
	Context("ValidateIP", func() {
		It("returns true when given a valid (IPv4, IPv6) address", func() {
			ipv4 := "4.4.4.4"
			ipv6 := "2001:db8::68"

			Expect(ValidateIP(ipv4)).To(BeTrue())
			Expect(ValidateIP(ipv6)).To(BeTrue())
		})

		It("returns false when given a invalid (IPv4, IPv6) address", func() {
			invalidIPv4 := []string{"256.256.256.256", "-1.1.1.1", "", "4", "2.2", "13.37.h4.x0rz"}
			for _, ip := range invalidIPv4 {
				Expect(ValidateIP(ip)).To(BeFalse())
			}

			// Source: http://www.ronnutter.com/ipv6-cheatsheet-on-identifying-valid-ipv6-addresses/
			invalidIPv6 := []string{"1200::AB00:1234::2552:7777:1313", "1200:0000:AB00:1234:O000:2552:7777:1313", ""}
			for _, ip := range invalidIPv6 {
				Expect(ValidateIP(ip)).To(BeFalse())
			}
		})
	})

	Context("IsResolvconf", func() {
		BeforeEach(func() {
			err := installResolvConf()
			Expect(err).NotTo(HaveOccurred())

			// Purging is required because otherwise it is not possible for this tool
			// to detect which implementation should be used. It is not possible to
			// have both implementations installed on a given system.

			err = purgePackage("openresolv")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := umount("/etc")
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

		It("updates the head", func() {
			cli, err := gexec.Build("github.com/cloudfoundry-incubator/resolvconf-manager")
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(cli, "-head", "99.99.99.99")
			session, err := gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			dat, err := ioutil.ReadFile("/etc/resolv.conf")
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				dat, err = ioutil.ReadFile("/etc/resolv.conf")
				Expect(err).NotTo(HaveOccurred())
				return serverIsFirst("99.99.99.99", string(dat))
			}, 10*time.Second).Should(BeTrue())

			cmd = exec.Command(cli, "-head", "98.98.98.98")
			session, err = gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Eventually(func() bool {
				dat, err = ioutil.ReadFile("/etc/resolv.conf")
				Expect(err).NotTo(HaveOccurred())
				return serverIsFirst("98.98.98.98", string(dat))
			}, 10*time.Second).Should(BeTrue())
		})

		It("updates the base", func() {
			cli, err := gexec.Build("github.com/cloudfoundry-incubator/resolvconf-manager")
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(cli, "-base", "1.2.3.4,4.5.6.7")
			session, err := gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			dat, err := ioutil.ReadFile("/etc/resolv.conf")
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				dat, err = ioutil.ReadFile("/etc/resolv.conf")
				Expect(err).NotTo(HaveOccurred())
				return baseMatches([]string{"1.2.3.4", "4.5.6.7"}, string(dat))
			}, 10*time.Second).Should(BeTrue())
		})

		It("updates both the head and base", func() {
			cli, err := gexec.Build("github.com/cloudfoundry-incubator/resolvconf-manager")
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(cli, "-head", "98.98.98.98", "-base", "1.2.3.4,4.5.6.7")
			session, err := gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			dat, err := ioutil.ReadFile("/etc/resolv.conf")
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				dat, err = ioutil.ReadFile("/etc/resolv.conf")
				Expect(err).NotTo(HaveOccurred())
				return serverIsFirst("98.98.98.98", string(dat)) &&
					baseMatches([]string{"1.2.3.4", "4.5.6.7"}, string(dat))
			}, 10*time.Second).Should(BeTrue())
		})

		It("errors if no input is provided", func() {
			cli, err := gexec.Build("github.com/cloudfoundry-incubator/resolvconf-manager")
			Expect(err).NotTo(HaveOccurred())
			cmd := exec.Command(cli)
			session, err := gexec.Start(cmd, ioutil.Discard, ioutil.Discard)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("IsOpenresolv", func() {
		BeforeEach(func() {
			err := installPackage("openresolv")
			Expect(err).NotTo(HaveOccurred())
			err = purgePackage("resolvconf")
			Expect(err).NotTo(HaveOccurred())
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
