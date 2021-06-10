// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package workflow

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/networkchaos"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/timechaos"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func TestcaseWorkflowSerial(
	ns string,
	kubeCli kubernetes.Interface,
	cli client.Client,
	c http.Client,
	port uint16,
	workloadLabels map[string]string,
) {
	const workflowE2eSerial = "workflow-e2e-serial"

	// podchaos for 20s -> sleep 10s -> timechaos for 20s -> sleep 10s -> done
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("wait e2e helper ready")
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	By("create the workflow")

	var pods *corev1.PodList
	var newPods *corev1.PodList
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(workloadLabels).String(),
	}
	pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
	Expect(err).ShouldNot(HaveOccurred())

	timeWhenWorkflowCreate := time.Now()
	const sleepDuration = 10 * time.Second
	const podChaosDuration = 20 * time.Second
	const timeChaosDuration = 20 * time.Second
	workflowSpec := commonSerialWorkflow(sleepDuration, podChaosDuration, timeChaosDuration, ns, workloadLabels)
	err = cli.Create(ctx, &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      workflowE2eSerial,
		},
		Spec: workflowSpec,
	})
	Expect(err).ShouldNot(HaveOccurred())

	// time skew chaos
	Eventually(func() bool {
		framework.Logf("assertion that time chaos is affected")
		podTimeNS, err := timechaos.GetPodTimeNS(c, port)
		if err != nil {
			By(fmt.Sprintf("failed to fetch time from pods, %s", err.Error()))
			return false
		}
		return time.Now().Sub(*podTimeNS).Round(time.Hour) == time.Hour
	}, "10s", "1s").Should(BeTrue())
	timeWhenTimeSkewChaosAffected := time.Now()
	By(fmt.Sprintf("time chaos in workflow affected, in %s", timeWhenTimeSkewChaosAffected.Sub(timeWhenWorkflowCreate)))

	// waiting for the recover of time skew chaos
	Eventually(func() bool {
		By("assertion that time chaos will be deleted")
		allTimeChaos := v1alpha1.TimeChaosList{}
		err := cli.List(ctx, &allTimeChaos, &client.ListOptions{Namespace: ns})
		if err != nil {
			By(fmt.Sprintf("failed to list time chaos, %s", err.Error()))
			return false
		}
		return len(allTimeChaos.Items) == 0
	}, "40s", "2s").Should(BeTrue())
	timeWhenTimeChaosRecovered := time.Now()

	By(fmt.Sprintf("time chaos in workflow recovered, in %s", timeWhenTimeChaosRecovered.Sub(timeWhenWorkflowCreate)))

	Eventually(func() bool {
		By("assertion that time skew should be recovered")
		framework.Logf("assertion that time chaos is affected")
		podTimeNS, err := timechaos.GetPodTimeNS(c, port)
		if err != nil {
			By(fmt.Sprintf("failed to fetch time from pods, %s", err.Error()))
			return false
		}
		return time.Now().Sub(*podTimeNS).Round(time.Hour) == 0
	}, "5s", "1s").Should(BeTrue())

	// waiting for the pod chaos
	Eventually(func() bool {
		framework.Logf("assertion that pod chaos is affected")
		newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			By(fmt.Sprintf("failed to list new pods, %s", err.Error()))
			return false
		}
		return !fixture.HaveSameUIDs(pods.Items, newPods.Items)
	}, "30s", "1s").Should(BeTrue())
	timeWhenFirstChaosAffected := time.Now()

	By(fmt.Sprintf("pod chaos in workflow affected, in %s", timeWhenFirstChaosAffected.Sub(timeWhenWorkflowCreate)))

	// waiting for the pod chaos
	Eventually(func() bool {
		By("assertion that pod chaos will be deleted")
		allPodChaos := v1alpha1.PodChaosList{}
		err := cli.List(ctx, &allPodChaos, &client.ListOptions{Namespace: ns})
		if err != nil {
			By(fmt.Sprintf("failed to list pod chaos, %s", err.Error()))
			return false
		}
		return len(allPodChaos.Items) == 0
	}, "40s", "2s").Should(BeTrue())
	timeWhenFirstChaosRecovered := time.Now()

	By(fmt.Sprintf("pod chaos in workflow recovered, in %s", timeWhenFirstChaosRecovered.Sub(timeWhenWorkflowCreate)))

	lastPods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
	Expect(err).ShouldNot(HaveOccurred())

	Consistently(func() bool {
		By("assertion that pod chaos will be recovered")
		latest, err := kubeCli.CoreV1().Pods(ns).List(listOption)
		if err != nil {
			By(fmt.Sprintf("failed to list new pods, %s", err.Error()))
			return false
		}
		defer func() {
			lastPods = latest
		}()
		return fixture.HaveSameUIDs(lastPods.Items, latest.Items)
	}, "5s", "1s").Should(BeTrue())

}

// it will kill all the pod, and inject -1h time skew for all pods
func commonSerialWorkflow(
	sleepDuration, podChaosDuration, timeChaosDuration time.Duration,
	ns string,
	workloadLabels map[string]string,
) v1alpha1.WorkflowSpec {
	const entry = "the-serial"
	const sleeping = "sleep-for-a-while"
	const podchaos = "pod-chaos"
	const timechaos = "time-chaos"

	sleepDurationString := sleepDuration.String()
	podChaosDurationString := podChaosDuration.String()
	timeChaosDurationString := timeChaosDuration.String()

	return v1alpha1.WorkflowSpec{
		Entry: entry,
		Templates: []v1alpha1.Template{
			{
				Name:     entry,
				Type:     v1alpha1.TypeSerial,
				Duration: nil,
				Tasks: []string{
					timechaos,
					sleeping,
					podchaos,
					sleeping,
				},
				EmbedChaos: nil,
			},
			{
				Name:       sleeping,
				Type:       v1alpha1.TypeSuspend,
				Duration:   &sleepDurationString,
				Tasks:      nil,
				EmbedChaos: nil,
			},
			{
				Name:     podchaos,
				Type:     v1alpha1.TypePodChaos,
				Duration: &podChaosDurationString,
				Tasks:    nil,
				EmbedChaos: &v1alpha1.EmbedChaos{
					PodChaos: &v1alpha1.PodChaosSpec{
						ContainerSelector: v1alpha1.ContainerSelector{
							PodSelector: v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									Namespaces:     []string{ns},
									LabelSelectors: workloadLabels,
								},
								Mode: v1alpha1.AllPodMode,
							},
						},
						Action: v1alpha1.PodKillAction,
					},
				},
			}, {
				Name:     timechaos,
				Type:     v1alpha1.TypeTimeChaos,
				Duration: &timeChaosDurationString,
				EmbedChaos: &v1alpha1.EmbedChaos{
					TimeChaos: &v1alpha1.TimeChaosSpec{
						ContainerSelector: v1alpha1.ContainerSelector{
							PodSelector: v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									Namespaces:     []string{ns},
									LabelSelectors: workloadLabels,
								},
								Mode: v1alpha1.AllPodMode,
							},
						},
						TimeOffset: "-1h",
					},
				},
			},
		},
	}
}

func TestcaseDeadlineOfSerial(
	ns string,
	kubeCli kubernetes.Interface,
	cli client.Client,
	c http.Client,
	peers []*corev1.Pod,
	ports []uint16,
	workloadLabels map[string]string,
) {
	const timeChaosShouldNotSpawned = "time-chaos-should-not-spawned"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("wait e2e helper ready")
	for _, port := range ports {
		err := util.WaitE2EHelperReady(c, port)
		framework.ExpectNoError(err, "wait e2e helper ready error")
	}

	By("create the workflow")

	timeWhenWorkflowCreate := time.Now()
	const serialDeadline = 10 * time.Second
	const networkChaosDuration = 20 * time.Second
	const timeChaosDuration = 40 * time.Second

	// network partition will be set:
	// partition 1: peer-0
	// partition 2: peer-1, peer-2, peer-3
	workflowSpec := secondOneShouldNotSpawned(serialDeadline, networkChaosDuration, timeChaosDuration, ns, workloadLabels)
	err := cli.Create(ctx, &v1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      timeChaosShouldNotSpawned,
		},
		Spec: workflowSpec,
	})
	Expect(err).ShouldNot(HaveOccurred())

	// assert that network chaos applied
	Eventually(func() bool {
		framework.Logf("assertion that network chaos is affected")
		conditions := networkchaos.ProbeNetworkCondition(c, peers, ports, false)
		blocked := conditions[networkchaos.NetworkConditionBlocked]
		By(fmt.Sprintf("blocked %+v", blocked))
		return len(blocked) == 3
	}, "10s", "1s").Should(BeTrue())

	timeWhenTimeSkewChaosAffected := time.Now()
	By(fmt.Sprintf("network chaos in workflow affected, in %s", timeWhenTimeSkewChaosAffected.Sub(timeWhenWorkflowCreate)))

	// assert that network chaos recovered
	Eventually(func() bool {
		framework.Logf("assertion that network chaos is recovered")
		conditions := networkchaos.ProbeNetworkCondition(c, peers, ports, false)
		blocked := conditions[networkchaos.NetworkConditionBlocked]
		By(fmt.Sprintf("blocked %+v", blocked))
		return len(blocked) == 0
	}, "30s", "1s", "asdasd").Should(BeTrue())
	timeWhenTimeChaosRecovered := time.Now()

	By(fmt.Sprintf("network chaos in workflow recovered, in %s", timeWhenTimeChaosRecovered.Sub(timeWhenWorkflowCreate)))

	// assert that time chaos should not be spawned
	Consistently(func() bool {
		framework.Logf("assertion that time chaos would never applied")
		podTimeNS, err := timechaos.GetPodTimeNS(c, ports[0])
		if err != nil {
			By(fmt.Sprintf("failed to fetch time from pods, %s", err.Error()))
			return false
		}
		return time.Now().Sub(*podTimeNS).Round(time.Hour) == 0
	}, "20s", "1s").Should(BeTrue())

}

func secondOneShouldNotSpawned(
	serialDeadline, networkChaosDuration, timeChaosDuration time.Duration,
	ns string,
	workloadLabels map[string]string,
) v1alpha1.WorkflowSpec {
	const entry = "the-serial"
	const networkChaos = "network-chaos"
	const timeChaos = "time-chaos"

	if serialDeadline > networkChaosDuration {
		panic("the deadline ")
	}

	deadlineString := serialDeadline.String()
	networkChaosDurationString := networkChaosDuration.String()
	timeChaosDurationString := timeChaosDuration.String()
	return v1alpha1.WorkflowSpec{
		Entry: entry,
		Templates: []v1alpha1.Template{
			{
				Name:     entry,
				Type:     v1alpha1.TypeSerial,
				Duration: &deadlineString,
				Tasks: []string{
					networkChaos,
					timeChaos,
				},
			}, {
				Name:     networkChaos,
				Type:     v1alpha1.TypeNetworkChaos,
				Duration: &networkChaosDurationString,
				EmbedChaos: &v1alpha1.EmbedChaos{NetworkChaos: &v1alpha1.NetworkChaosSpec{
					PodSelector: v1alpha1.PodSelector{
						Selector: v1alpha1.PodSelectorSpec{
							LabelSelectors: map[string]string{
								"app": "network-peer-0",
							},
						},
						Mode: v1alpha1.AllPodMode,
					},
					Action:      v1alpha1.PartitionAction,
					Duration:    &networkChaosDurationString,
					TcParameter: v1alpha1.TcParameter{},
					Target: &v1alpha1.PodSelector{
						Selector: v1alpha1.PodSelectorSpec{
							Namespaces: []string{
								ns,
							},
						},
						Mode: v1alpha1.AllPodMode,
					},
				}},
			}, {
				Name:     timeChaos,
				Type:     v1alpha1.TypeTimeChaos,
				Duration: &timeChaosDurationString,
				EmbedChaos: &v1alpha1.EmbedChaos{
					TimeChaos: &v1alpha1.TimeChaosSpec{
						ContainerSelector: v1alpha1.ContainerSelector{
							PodSelector: v1alpha1.PodSelector{
								Selector: v1alpha1.PodSelectorSpec{
									Namespaces:     []string{ns},
									LabelSelectors: workloadLabels,
								},
								Mode: v1alpha1.AllPodMode,
							},
						},
						TimeOffset: "-1h",
					},
				},
			},
		},
	}
}
