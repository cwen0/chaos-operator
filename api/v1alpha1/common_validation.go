// Copyright 2020 PingCAP, Inc.
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
	"strconv"
	"time"

	cronv3 "github.com/robfig/cron/v3"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// ValidateSchedulerError defines the error message for ValidateScheduler
	ValidateSchedulerError = "schedule and duration should be omitted or defined at the same time"

	// ValidatePodchaosSchedulerError defines the error message for ValidateScheduler of Podchaos
	ValidatePodchaosSchedulerError = "schedule should be omitted"

	// ValidateValueParseError defines the error message for value parse error
	ValidateValueParseError = "parse value field error:%s"
)

// ValidateScheduler validates the InnerSchedulerObject
func ValidateScheduler(schedulerObject InnerSchedulerObject, spec *field.Path) field.ErrorList {

	allErrs := field.ErrorList{}

	nextStart := schedulerObject.GetNextStart()
	nextRecover := schedulerObject.GetNextRecover()
	if nextStart.After(nextRecover) {
		nextStartField := spec.Child("nextStart")
		allErrs = append(allErrs, field.Invalid(nextStartField, nextStart,
			fmt.Sprintf("the nextStart: \"%s\" can not after the nextRecover: \"%s\"", nextStart, nextRecover)))
	}

	schedulerField := spec.Child("scheduler")
	durationField := spec.Child("duration")
	duration, err := schedulerObject.GetDuration()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(durationField, nil,
			fmt.Sprintf("parse duration field error:%s", err)))
	}

	scheduler := schedulerObject.GetScheduler()

	if duration != nil && scheduler != nil {
		errs := validateSchedulerParams(duration, durationField, scheduler, schedulerField)
		if len(errs) != 0 {
			allErrs = append(allErrs, errs...)
		}
	} else if (duration == nil && scheduler != nil) || (duration != nil && scheduler == nil) {
		allErrs = append(allErrs, field.Invalid(schedulerField, scheduler, ValidateSchedulerError))
	}
	return allErrs
}

func validateSchedulerParams(duration *time.Duration, durationField *field.Path, spec *SchedulerSpec, schedulerField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if duration != nil && spec != nil {
		cronField := schedulerField.Child("cron")
		allErrs = append(validateCron(spec.Cron, cronField))

		scheduler, _ := cronv3.ParseStandard(spec.Cron)
		if scheduler != nil {
			tmpTime := time.Time{}
			nextTime := scheduler.Next(tmpTime)
			interval := nextTime.Sub(tmpTime)
			if *duration >= interval {
				allErrs = append(allErrs, field.Invalid(cronField, spec.Cron,
					fmt.Sprintf("the scheduling interval:\"%s\" must be greater than the duration:%s", spec.Cron, *duration)))
			}
		}
	}
	return allErrs
}
func validateCron(cron string, cronField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, err := cronv3.ParseStandard(cron)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(cronField,
			cron,
			fmt.Sprintf("parse cron field error:%s", err)))
	}
	return allErrs
}

// ValidatePodMode validates the value with podmode
func ValidatePodMode(value string, mode PodMode, valueField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	switch mode {
	case FixedPodMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf(ValidateValueParseError, err)))
		}

		if num <= 0 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value must be greater than 0 with mode:%s", FixedPodMode)))
		}

	case RandomMaxPercentPodMode, FixedPercentPodMode:
		percentage, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf(ValidateValueParseError, err)))
		}

		if percentage <= 0 || percentage > 100 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value of %d is invalid, Must be (0,100] with mode:%s",
					percentage, mode)))
		}
	}
	return allErrs
}
