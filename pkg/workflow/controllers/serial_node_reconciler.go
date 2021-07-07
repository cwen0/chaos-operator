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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

// SerialNodeReconciler watches on nodes which type is Serial
type SerialNodeReconciler struct {
	*ChildNodesFetcher
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	logger        logr.Logger
}

func NewSerialNodeReconciler(kubeClient client.Client, eventRecorder recorder.ChaosRecorder, logger logr.Logger) *SerialNodeReconciler {
	return &SerialNodeReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logger),
		kubeClient:        kubeClient,
		eventRecorder:     eventRecorder,
		logger:            logger,
	}
}

// Reconcile should be invoked by: changes on a serial node, or changes on a node which controlled by serial node.
// So we need to setup EnqueueRequestForOwner while setting up this reconciler.
//
// Reconcile does these things:
// 1. walk through on tasks in spec, compare them with the node instances (listed with v1alpha1.LabelControlledBy),
// remove the outdated instance;
// 2. find out the node needs to be created, then create one if exists;
// 3. update the status of serial node;
//
// In this reconciler, we SHOULD NOT use v1alpha1.WorkflowNodeStatus as the state.
// Because v1alpha1.WorkflowNodeStatus is generated by this reconciler, if that itself also depends on that state,
// it will be complex to decide when to update the status, and even require to update status more than one time,
// that sounds not good.
// And We MUST update v1alpha1.WorkflowNodeStatus by "observing real world" at EACH TIME, such as listing controlled
// children nodes.
// We only update v1alpha1.WorkflowNodeStatus once(wrapped with retry on conflict), at the end of this method.
func (it *SerialNodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	startTime := time.Now()
	defer func() {
		it.logger.V(4).Info("Finished syncing for serial node",
			"node", request.NamespacedName,
			"duration", time.Since(startTime),
		)
	}()

	ctx := context.TODO()

	node := v1alpha1.WorkflowNode{}
	err := it.kubeClient.Get(ctx, request.NamespacedName, &node)
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// only resolve serial nodes
	if node.Spec.Type != v1alpha1.TypeSerial {
		return reconcile.Result{}, nil
	}

	it.logger.V(4).Info("resolve serial node", "node", request)

	// make effects, create/remove children nodes
	err = it.syncChildNodes(ctx, node)
	if err != nil {
		return reconcile.Result{}, err
	}

	// update status
	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeNeedUpdate := v1alpha1.WorkflowNode{}
		err := it.kubeClient.Get(ctx, request.NamespacedName, &nodeNeedUpdate)
		if err != nil {
			return err
		}

		activeChildren, finishedChildren, err := it.fetchChildNodes(ctx, nodeNeedUpdate)
		if err != nil {
			return err
		}

		nodeNeedUpdate.Status.FinishedChildren = nil
		for _, finishedChild := range finishedChildren {
			nodeNeedUpdate.Status.FinishedChildren = append(nodeNeedUpdate.Status.FinishedChildren,
				corev1.LocalObjectReference{
					Name: finishedChild.Name,
				})
		}

		nodeNeedUpdate.Status.ActiveChildren = nil
		for _, activeChild := range activeChildren {
			nodeNeedUpdate.Status.ActiveChildren = append(nodeNeedUpdate.Status.ActiveChildren,
				corev1.LocalObjectReference{
					Name: activeChild.Name,
				})
		}

		if len(activeChildren) > 1 {
			it.logger.Info("warning: serial node has more than 1 active children", "namespace", nodeNeedUpdate.Namespace, "name", nodeNeedUpdate.Name, "children", nodeNeedUpdate.Status.ActiveChildren)
		}

		// TODO: also check the consistent between spec in task and the spec in child node
		if len(finishedChildren) == len(nodeNeedUpdate.Spec.Children) {
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionTrue,
				Reason: "",
			})
			it.eventRecorder.Event(&nodeNeedUpdate,recorder.NodeAccomplished{})
		} else {
			SetCondition(&nodeNeedUpdate.Status, v1alpha1.WorkflowNodeCondition{
				Type:   v1alpha1.ConditionAccomplished,
				Status: corev1.ConditionFalse,
				Reason: "",
			})
		}

		return it.kubeClient.Status().Update(ctx, &nodeNeedUpdate)
	})

	if updateError != nil {
		it.logger.Error(err, "failed to update the status of node", "node", request)
		return reconcile.Result{}, updateError
	}

	return reconcile.Result{}, nil
}

// syncChildNodes reconciles the children nodes to following the desired states.
// It does the first 2 steps mentioned in Reconcile.
//
// Notice again: we SHOULD NOT decide the operation based on v1alpha1.WorkflowNodeStatus, please
// use kubeClient to fetch information from real world.
func (it *SerialNodeReconciler) syncChildNodes(ctx context.Context, node v1alpha1.WorkflowNode) error {

	// empty serial node
	if len(node.Spec.Children) == 0 {
		it.logger.V(4).Info("empty serial node, NOOP",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		)
		return nil
	}

	if WorkflowNodeFinished(node.Status) {
		return nil
	}

	activeChildNodes, finishedChildNodes, err := it.fetchChildNodes(ctx, node)
	if err != nil {
		return err
	}
	var taskToStartup string
	if len(activeChildNodes) == 0 {
		// no active children, trying to spawn a new one
		for index, task := range node.Spec.Children {
			// Walking through on the Spec.Children, each one of task SHOULD has one corresponding workflow node;
			// If the spec of one task has been changed, the corresponding workflow node and other
			// workflow nodes **behinds** that workflow node will be deleted.
			// That's so called "partial rerun" feature.
			// For example:
			// One serial node have three children nodes: A, B, C, and all of them have finished.
			// Then user updates the Spec.Children[B], the expected behavior is workflow node B and C will be
			// deleted, then create a new node that refs to B, no effects on A.
			if index < len(finishedChildNodes) {
				// TODO: if the definition/spec of task changed, we should also respawn the node
				// child node start with task name

				// TODO: maybe the changes on Spec.Children should be concerned each time, not only during spawning
				// new instances, for shutdown outdated nodes **instantly**

				if strings.HasPrefix(task, finishedChildNodes[index].Name) {
					// TODO: emit event
					taskToStartup = task

					// TODO: nodes to delete should be all other unrecognized children nodes, include not contained in finishedChildNodes
					// delete that related nodes with best-effort pattern
					nodesToDelete := finishedChildNodes[index:]
					for _, refToDelete := range nodesToDelete {
						nodeToDelete := v1alpha1.WorkflowNode{}
						err := it.kubeClient.Get(ctx, types.NamespacedName{
							Namespace: node.Namespace,
							Name:      refToDelete.Name,
						}, &nodeToDelete)
						if client.IgnoreNotFound(err) != nil {
							it.logger.Error(err, "failed to fetch outdated child node",
								"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
								"child node", fmt.Sprintf("%s/%s", node.Namespace, nodeToDelete.Name))
						}
						err = it.kubeClient.Delete(ctx, &nodeToDelete)
						if client.IgnoreNotFound(err) != nil {
							it.logger.Error(err, "failed to fetch outdated child node",
								"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
								"child node", fmt.Sprintf("%s/%s", node.Namespace, nodeToDelete.Name))
						}
					}
					break
				}
			} else {
				// spawn child node
				taskToStartup = task
				break
			}
		}
	} else {
		it.logger.V(4).Info("serial node has active child/children, skip scheduling",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			"active children", activeChildNodes)
	}

	if len(taskToStartup) == 0 {
		it.logger.Info("no need to spawn new child node", "node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return nil
	}

	parentWorkflow := v1alpha1.Workflow{}
	err = it.kubeClient.Get(ctx, types.NamespacedName{
		Namespace: node.Namespace,
		Name:      node.Spec.WorkflowName,
	}, &parentWorkflow)
	if err != nil {
		it.logger.Error(err, "failed to fetch parent workflow",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
			"workflow name", node.Spec.WorkflowName)
		return err
	}
	// TODO: using ordered id instead of random suffix is better, like StatefulSet, also related to the sorting
	childNodes, err := renderNodesByTemplates(&parentWorkflow, &node, taskToStartup)
	if err != nil {
		it.logger.Error(err, "failed to render children childNodes",
			"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name))
		return err
	}

	var childrenNames []string
	for _, childNode := range childNodes {
		err := it.kubeClient.Create(ctx, childNode)
		if err != nil {
			it.logger.Error(err, "failed to create child node",
				"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
				"child node", childNode)
			return err
		}
		childrenNames = append(childrenNames, childNode.Name)
	}
	it.eventRecorder.Event(&node, recorder.NodesCreated{ChildNodes: childrenNames})
	it.logger.Info("serial node spawn new child node",
		"node", fmt.Sprintf("%s/%s", node.Namespace, node.Name),
		"child node", childrenNames)

	return nil
}
