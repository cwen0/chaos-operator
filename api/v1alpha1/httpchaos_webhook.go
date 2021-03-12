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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var httpchaoslog = logf.Log.WithName("httpchaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-awschaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=awschaos,verbs=create;update,versions=v1alpha1,name=mawschaos.kb.io

var _ webhook.Defaulter = &HTTPChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *HTTPChaos) Default() {
	httpchaoslog.Info("default", "name", in.Name)
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-awschaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=awschaos,versions=v1alpha1,name=vawschaos.kb.io

var _ ChaosValidator = &HTTPChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateCreate() error {
	httpchaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateUpdate(old runtime.Object) error {
	httpchaoslog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateDelete() error {
	httpchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *HTTPChaos) Validate() error {
	specField := field.NewPath("spec")
	allErrs := in.ValidateScheduler(specField)
	allErrs = append(allErrs, in.ValidatePodMode(specField)...)

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *HTTPChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	return ValidateScheduler(in, spec)
}

// ValidatePodMode validates the value with podmode
func (in *HTTPChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	// Because aws chaos does not need a pod mode, so return nil here.
	return nil
}

// SelectSpec returns the selector config for authority validate
func (in *HTTPChaos) GetSelectSpec() []SelectSpec {
	return nil
}
