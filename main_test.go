package main_test

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/resolvconf-manager"
)

const DeviceResourceBusy = "Device or resource busy"
const UnableLocatePackage = "Unable to locate package"

var _ = Describe("Main", func() {
	BeforeEach(func() {
		Mount()
	})

	Context("IsResolvconf", func() {
		BeforeEach(func() {
			InstallPackage("resolvconf")
			PurgePackage("openresolv")
		})

		It("detects the resolvconf package is installed", func() {
			Expect(IsResolvconf()).To(BeTrue())
			Expect(IsOpenresolv()).To(BeFalse())
		})

		// Do stuff
	})

	Context("IsOpenresolv", func() {
		BeforeEach(func() {
			InstallPackage("openresolv")
			PurgePackage("resolvconf")
		})

		It("detects the openresolv package is installed", func() {
			Expect(IsResolvconf()).To(BeFalse())
			Expect(IsOpenresolv()).To(BeTrue())
		})

		// Do stuff
	})

	AfterEach(func() {
		Unmount()
	})
})

func PurgePackage(pkg string) {
	fmt.Printf("Purging %s...\n", pkg)
	cmd := exec.Command("apt", "purge", "-y", pkg)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil && !IsKnownError(string(output)) {
		log.Fatal(err)
	}
}

func IsKnownError(output string) bool {
	return strings.Contains(output, UnableLocatePackage) ||
		strings.Contains(output, DeviceResourceBusy)
}

func InstallPackage(pkg string) {
	Unmount()

	err := exec.Command("apt", "update").Run()
	if err != nil {
		log.Fatal(err)
	}

	// When installing resolvconf or openresolv, we expect to see an error trying
	// to create a symlink to /etc/resolv.conf. The error is usually "Device or
	// resource busy".

	fmt.Printf("Installing %s...", pkg)
	cmd := exec.Command("apt", "install", "-y", pkg)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil && !IsKnownError(string(output)) {
		log.Fatal(err)
	}

	Mount()
}

func Mount() {
	fmt.Println("Mounting /etc/resolv.conf...")

	cmd := exec.Command("bash", "/mount-etc-resolv.conf.sh")
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		log.Fatal(err)
	}
}

func Unmount() {
	fmt.Println("Unmounting /etc/resolv.conf...")

	cmd := exec.Command("umount", "/etc/resolv.conf")
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		log.Fatal(err)
	}
}
