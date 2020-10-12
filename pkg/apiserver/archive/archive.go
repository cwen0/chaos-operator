// Copyright 2020 Chaos Mesh Authors.
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

package archive

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Service defines a handler service for archive experiments.
type Service struct {
	conf    *config.ChaosDashboardConfig
	kubeCli client.Client
	archive core.ExperimentStore
	event   core.EventStore
}

// NewService returns an archive experiment service instance.
func NewService(
	conf *config.ChaosDashboardConfig,
	cli client.Client,
	archive core.ExperimentStore,
	event core.EventStore,
) *Service {
	return &Service{
		conf:    conf,
		kubeCli: cli,
		archive: archive,
		event:   event,
	}
}

// Register mounts our HTTP handler on the mux.
func Register(r *gin.RouterGroup, s *Service) {
	endpoint := r.Group("/archives")

	endpoint.GET("", s.list)
	endpoint.GET("/detail", s.detail)
	endpoint.GET("/report", s.report)
}

// Detail represents an archive instance.
type Detail struct {
	core.Archive
	YAML core.ExperimentYAMLDescription `json:"yaml"`
}

// Report defines the report of archive experiments.
type Report struct {
	Meta           *core.Archive `json:"meta"`
	Events         []*core.Event `json:"events"`
	TotalTime      string        `json:"total_time"`
	TotalFaultTime string        `json:"total_fault_time"`
}

// @Summary Get archived chaos experiments.
// @Description Get archived chaos experiments.
// @Tags archives
// @Produce json
// @Param namespace query string false "namespace"
// @Param name query string false "name"
// @Param kind query string false "kind" Enums(PodChaos, IoChaos, NetworkChaos, TimeChaos, KernelChaos, StressChaos)
// @Success 200 {array} core.Archive
// @Router /archives [get]
// @Failure 500 {object} utils.APIError
func (s *Service) list(c *gin.Context) {
	kind := c.Query("kind")
	name := c.Query("name")
	ns := c.Query("namespace")

	data, err := s.archive.ListMeta(context.Background(), kind, ns, name)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		return
	}

	var archives []core.Archive

	for _, d := range data {
		archives = append(archives, core.Archive{
			UID:        d.UID,
			Kind:       d.Kind,
			Namespace:  d.Namespace,
			Name:       d.Name,
			Action:     d.Action,
			StartTime:  d.StartTime,
			FinishTime: d.FinishTime,
		})
	}

	c.JSON(http.StatusOK, data)
}

// @Summary Get the detail of archived chaos experiment.
// @Description Get the detail of archived chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {object} Detail
// @Router /archives/detail [get]
// @Failure 500 {object} utils.APIError
func (s *Service) detail(c *gin.Context) {
	var (
		err    error
		yaml   core.ExperimentYAMLDescription
		detail Detail
	)
	uid := c.Query("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	data, err := s.archive.FindByUID(context.TODO(), uid)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		}
		return
	}

	switch data.Kind {
	case v1alpha1.KindPodChaos:
		yaml, err = data.ParsePodChaos()
	case v1alpha1.KindIOChaos:
		yaml, err = data.ParseIOChaos()
	case v1alpha1.KindNetworkChaos:
		yaml, err = data.ParseNetworkChaos()
	case v1alpha1.KindTimeChaos:
		yaml, err = data.ParseTimeChaos()
	case v1alpha1.KindKernelChaos:
		yaml, err = data.ParseKernelChaos()
	case v1alpha1.KindStressChaos:
		yaml, err = data.ParseStressChaos()
	default:
		err = fmt.Errorf("kind %s is not support", data.Kind)
	}
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.WrapWithNoMessage(err))
		return
	}

	detail = Detail{
		Archive: core.Archive{
			UID:        data.UID,
			Kind:       data.Kind,
			Name:       data.Name,
			Namespace:  data.Namespace,
			Action:     data.Action,
			StartTime:  data.StartTime,
			FinishTime: data.FinishTime,
		},
		YAML: yaml,
	}

	c.JSON(http.StatusOK, detail)
}

// @Summary Get the report of a chaos experiment.
// @Description Get the report of a chaos experiment.
// @Tags archives
// @Produce json
// @Param uid query string true "uid"
// @Success 200 {array} Report
// @Router /archives/report [get]
// @Failure 500 {object} utils.APIError
func (s *Service) report(c *gin.Context) {
	var (
		err       error
		report    Report
	)
	uid := c.Query("uid")

	if uid == "" {
		c.Status(http.StatusBadRequest)
		_ = c.Error(utils.ErrInvalidRequest.New("uid cannot be empty"))
		return
	}

	meta, err := s.archive.FindMetaByUID(context.Background(), uid)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		} else {
			c.Status(http.StatusInternalServerError)
			_ = c.Error(utils.ErrInvalidRequest.New("the archive is not found"))
		}
		return
	}
	report.Meta = &core.Archive{
		UID:        meta.UID,
		Kind:       meta.Kind,
		Namespace:  meta.Namespace,
		Name:       meta.Name,
		Action:     meta.Action,
		StartTime:  meta.StartTime,
		FinishTime: meta.FinishTime,
	}

	report.TotalTime = report.Meta.FinishTime.Sub(report.Meta.StartTime).String()

	report.Events, err = s.event.ListByUID(context.TODO(), uid)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(utils.ErrInternalServer.NewWithNoMessage())
		return
	}

	timeNow := time.Now()
	timeAfter := timeNow
	for _, et := range report.Events {
		timeAfter = timeAfter.Add(et.FinishTime.Sub(*et.StartTime))
	}
	report.TotalFaultTime = timeAfter.Sub(timeNow).String()

	c.JSON(http.StatusOK, report)
}
