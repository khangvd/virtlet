/*
Copyright 2017 Mirantis

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"regexp"
	"strconv"
	"time"

	. "github.com/onsi/gomega"

	"github.com/Mirantis/virtlet/tests/e2e/framework"
	. "github.com/Mirantis/virtlet/tests/e2e/ginkgo-ext"
)

var _ = Describe("VM resources", func() {
	var (
		vm  *framework.VMInterface
		ssh framework.Executor
	)

	BeforeAll(func() {
		vm = controller.VM("vm-resources")
		Expect(vm.CreateAndWait(VMOptions{
			VCPUCount:      2,
			RootVolumeSize: "4Gi",
		}.ApplyDefaults(), time.Minute*5, nil)).To(Succeed())
		do(vm.Pod())
	})

	scheduleWaitSSH(&vm, &ssh)

	AfterAll(func() {
		deleteVM(vm)
	})

	It("Should have CPU count as set for the domain [Conformance]", func() {
		checkCPUCount(vm, ssh, 2)
	})

	It("Should have total memory amount close to that set for the domain [Conformance]", func() {
		meminfo := do(framework.RunSimple(ssh, "cat", "/proc/meminfo")).(string)
		totals := regexp.MustCompile(`(?:DirectMap4k|DirectMap2M):\s+(\d+)`).FindAllStringSubmatch(meminfo, -1)
		Expect(totals).To(HaveLen(2))
		total := 0
		for _, m := range totals {
			Expect(m).To(HaveLen(2))
			total += do(strconv.Atoi(m[1])).(int)
		}
		Expect(total).To(BeNumerically(">", 1024*(*memoryLimit-1)))
		Expect(total).To(BeNumerically("<", 1024*(*memoryLimit)))
	})

	It("Should grow the root volume size if requested", func() {
		minSize := 3900
		Eventually(func() (int, error) {
			sizeStr := do(framework.RunSimple(ssh, "/bin/sh", "-c", `df -m / | tail -1 | awk "{print \$2}"`)).(string)
			size, err := strconv.Atoi(sizeStr)
			if err != nil {
				return 0, err
			}
			return size, nil
		}).Should(BeNumerically(">=", minSize))
	})
})
