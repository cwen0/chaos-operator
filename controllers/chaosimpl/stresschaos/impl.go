// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package stresschaos

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type Impl struct {
	client.Client

	Log logr.Logger

	decoder *utils.ContianerRecordDecoder
}

var _ common.ChaosImpl = (*Impl)(nil)

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index])
	pbClient := decodedContainer.PbClient
	containerId := decodedContainer.ContainerId
	if pbClient != nil {
		defer pbClient.Close()
	}

	if err != nil {
		return v1alpha1.NotInjected, err
	}

	stresschaos := obj.(*v1alpha1.StressChaos)
	if stresschaos.Status.Instances == nil {
		stresschaos.Status.Instances = make(map[string]v1alpha1.StressInstance)
	}
	_, ok := stresschaos.Status.Instances[records[index].Id]
	if ok {
		impl.Log.Info("an stress-ng instance is running for this pod")
		return v1alpha1.Injected, nil
	}

	stressors := stresschaos.Spec.StressngStressors
	cpuStressors := ""
	memoryStressors := ""
	if len(stressors) == 0 {
		cpuStressors, memoryStressors, err = stresschaos.Spec.Stressors.Normalize()
		if err != nil {
			impl.Log.Info("fail to ")
			// TODO: add an event here
			return v1alpha1.NotInjected, err
		}
	}

	res, err := pbClient.ExecStressors(ctx, &pb.ExecStressRequest{
		Scope:           pb.ExecStressRequest_CONTAINER,
		Target:          containerId,
		CpuStressors:    cpuStressors,
		MemoryStressors: memoryStressors,
		EnterNS:         true,
	})

	if err != nil {
		return v1alpha1.NotInjected, err
	}
	// TODO: support custom status
	impl.Log.Info("message", "cpuidd", res.CpuInstance, "memoryidd", res.MemoryInstance)
	stresschaos.Status.Instances[records[index].Id] = v1alpha1.StressInstance{
		CpuUID: res.CpuInstance,
		CpuStartTime: &metav1.Time{
			Time: time.Unix(res.CpuStartTime/1000, (res.CpuStartTime%1000)*int64(time.Millisecond)),
		},
		MemoryUID: res.MemoryInstance,
		MemoryStartTime: &metav1.Time{
			Time: time.Unix(res.MemoryStartTime/1000, (res.MemoryStartTime%1000)*int64(time.Millisecond)),
		},
	}
	
	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index])
	pbClient := decodedContainer.PbClient
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		if utils.IsFailToGet(err) {
			// pretend the disappeared container has been recovered
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	stresschaos := obj.(*v1alpha1.StressChaos)
	if stresschaos.Status.Instances == nil {
		return v1alpha1.NotInjected, nil
	}
	instance, ok := stresschaos.Status.Instances[records[index].Id]
	impl.Log.Info("message", "index", index, "id", records[index].Id)
	impl.Log.Info("message", "cpuid", instance.CpuStartTime, "memoryid", instance.MemoryStartTime)
	if !ok {
		impl.Log.Info("Pod seems already recovered", "pod", decodedContainer.Pod.UID)
		return v1alpha1.NotInjected, nil
	}
	if _, err = pbClient.CancelStressors(ctx, &pb.CancelStressRequest{
		CpuInstance:     instance.CpuUID,
		CpuStartTime:    instance.CpuStartTime.UnixNano() / int64(time.Millisecond),
		MemoryInstance:  instance.MemoryUID,
		MemoryStartTime: instance.MemoryStartTime.UnixNano() / int64(time.Millisecond),
	}); err != nil {
		// TODO: check whether the erorr still exists
		return v1alpha1.Injected, nil
	}
	delete(stresschaos.Status.Instances, records[index].Id)
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContianerRecordDecoder) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "stresschaos",
		Object: &v1alpha1.StressChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("stresschaos"),
			decoder: decoder,
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
