package main_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestResolvconfManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ResolvconfManager Suite")
}

func AptUpdate() error {
	dat, err := ioutil.ReadFile("/etc/resolv.conf")
	Expect(err).NotTo(HaveOccurred())
	fmt.Println(string(dat))
	cmd := exec.Command("apt", "update")
	ses, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Eventually(ses, 1*time.Minute).Should(gexec.Exit(0))
	if err != nil {
		return err
	}
	return nil
}

var _ = BeforeSuite(func() {
	err := AptUpdate()
	Expect(err).NotTo(HaveOccurred())
})
