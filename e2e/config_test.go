/*
Copyright 2022 The KubeVela Authors.
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
	"context"
	"os"
	"time"

	"github.com/kubevela/workflow/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/oam-dev/kubevela/pkg/oam/util"
)

var _ = Describe("Test the workflow run with the config steps", func() {
	ctx := context.Background()

	var namespace string
	var ns corev1.Namespace

	BeforeEach(func() {
		namespace = "config-e2e-test"
		ns = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}

		Eventually(func() error {
			return k8sClient.Create(ctx, &ns)
		}, time.Second*3, time.Microsecond*300).Should(SatisfyAny(BeNil(), &util.AlreadyExistMatcher{}))
	})

	It("Test the workflow", func() {
		content, err := os.ReadFile("./test-data/config-workflow-run.yaml")
		Expect(err).Should(BeNil())
		var workflowRun v1alpha1.WorkflowRun
		Expect(yaml.Unmarshal(content, &workflowRun)).Should(BeNil())
		workflowRun.Namespace = namespace
		Expect(k8sClient.Create(context.TODO(), &workflowRun)).Should(BeNil())
		Eventually(
			func() v1alpha1.WorkflowRunPhase {
				var getWorkflow v1alpha1.WorkflowRun
				if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: workflowRun.Name}, &getWorkflow); err != nil {
					klog.Errorf("fail to query the app %s", err.Error())
				}
				klog.Infof("the workflow run status is %s (%+v)", getWorkflow.Status.Phase, getWorkflow.Status.Steps)
				return getWorkflow.Status.Phase
			},
			time.Second*30, time.Second*2).Should(Equal(v1alpha1.WorkflowStateSucceeded))
	})

	AfterEach(func() {
		By("Clean up resources after a test")
		k8sClient.DeleteAllOf(ctx, &v1alpha1.WorkflowRun{}, client.InNamespace(namespace))
	})
})
