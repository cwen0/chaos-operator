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

package chaosdaemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pkg/errors"
)

const (
	todaBin                      = "/usr/local/bin/toda"
	todaUnixSocketFilePath       = "/toda.sock"
	todaClientUnixScoketFilePath = "/proc/%d/root/toda.sock"
)

func (s *DaemonServer) ApplyIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) (*pb.ApplyIOChaosResponse, error) {
	log := s.getLoggerFromContext(ctx)
	log.Info("applying io chaos", "Request", in)

	if in.InstanceUid == "" {
		if uid, ok := s.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: int(in.Instance), CreateTime: in.StartTime}); ok {
			in.InstanceUid = uid
		}
	}

	if in.InstanceUid != "" {
		if err := s.killIOChaos(ctx, in.InstanceUid); err != nil {
			// ignore this error
			log.Error(err, "kill background process", "uid", in.InstanceUid)
		}
	}

	var actions []v1alpha1.IOChaosAction
	err := json.Unmarshal([]byte(in.Actions), &actions)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal json bytes")
	}

	log.Info("the length of actions", "length", len(actions))
	if len(actions) == 0 {
		return &pb.ApplyIOChaosResponse{
			Instance:  0,
			StartTime: 0,
		}, nil
	}

	if err := s.createIOChaos(ctx, in); err != nil {
		return nil, errors.Wrap(err, "create IO chaos")
	}

	log.Info("Waiting for toda to start")
	resp, err := s.applyIOChaos(ctx, in)
	if err != nil {
		if kerr := s.killIOChaos(ctx, in.InstanceUid); kerr != nil {
			log.Error(kerr, "kill toda", "request", in)
		}
		return nil, errors.Wrap(err, "apply config")
	}
	return resp, err
}

func (s *DaemonServer) killIOChaos(ctx context.Context, uid string) error {
	log := s.getLoggerFromContext(ctx)

	err := s.backgroundProcessManager.KillBackgroundProcess(ctx, uid)
	if err != nil {
		return errors.Wrapf(err, "kill toda %s", uid)
	}
	log.Info("kill toda successfully", "uid", uid)
	return nil
}

func (s *DaemonServer) applyIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) (*pb.ApplyIOChaosResponse, error) {
	log := s.getLoggerFromContext(ctx)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		return nil, errors.Wrap(err, "getting PID")
	}

	transport := &unixSocketTransport{
		addr: fmt.Sprintf(todaClientUnixScoketFilePath, pid),
	}

	req, err := http.NewRequest(http.MethodPut, "http://psedo-host/update", bytes.NewReader([]byte(in.Actions)))
	if err != nil {
		return nil, errors.Wrap(err, "create http://psedo-host/update request")
	}

	err = retry.Do(func() error {
		_, retryErr := transport.RoundTrip(req)
		if retryErr != nil {
			log.Error(retryErr, "transport RoundTrip http://psedo-host/update")
			return errors.Wrap(retryErr, "transport RoundTrip http://psedo-host/update")
		}
		return nil
	}, retry.Delay(time.Second*5),
		retry.Attempts(2))

	if err != nil {
		log.Error(err, "applyIOChaos update io retry Do")
		return nil, err
	}

	req, err = http.NewRequest(http.MethodPut, "http://psedo-host/get_status", bytes.NewReader([]byte("ping")))
	if err != nil {
		return nil, errors.Wrap(err, "create http://psedo-host/get_status request")
	}

	err = retry.Do(func() error {
		resp, retryErr := transport.RoundTrip(req)
		if retryErr != nil {
			log.Error(retryErr, "retry send http request, error")
			return errors.Wrap(retryErr, "retry send http request")
		}
		body, retryErr := io.ReadAll(resp.Body)
		if retryErr != nil || string(body) != "ok" {
			log.Error(retryErr, "toda startup takes too long or an error occurs, error")
			return errors.Wrap(retryErr, "toda startup takes too long or an error occurs")
		}
		return nil
	}, retry.Delay(time.Second*5),
		retry.Attempts(2))

	if err != nil {
		log.Error(err, "applyIOChaos get status io retry Do")
		return nil, errors.Wrap(err, "send http request")
	}

	log.Info("io chaos applied")

	return &pb.ApplyIOChaosResponse{
		Instance:    in.Instance,
		StartTime:   in.StartTime,
		InstanceUid: in.InstanceUid,
	}, nil
}

func (s *DaemonServer) createIOChaos(ctx context.Context, in *pb.ApplyIOChaosRequest) error {
	log := s.getLoggerFromContext(ctx)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		return errors.Wrap(err, "getting PID")
	}

	// TODO: make this log level configurable
	args := fmt.Sprintf("--path %s --verbose info --interactive-path %s", in.Volume, todaUnixSocketFilePath)
	log.Info("executing", "cmd", todaBin+" "+args)

	processBuilder := bpm.DefaultProcessBuilder(todaBin, strings.Split(args, " ")...).
		EnableLocalMnt().
		SetIdentifier(fmt.Sprintf("toda-%s", in.ContainerId))

	if in.EnterNS {
		processBuilder = processBuilder.SetNS(pid, bpm.MountNS).SetNS(pid, bpm.PidNS)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := processBuilder.Build(ctx)
	cmd.Stderr = os.Stderr
	proc, err := s.backgroundProcessManager.StartProcess(ctx, cmd)
	if err != nil {
		return errors.Wrapf(err, "start process `%s`", cmd)
	}

	in.Instance = int64(proc.Pair.Pid)
	in.StartTime = proc.Pair.CreateTime
	in.InstanceUid = proc.Uid
	return nil

}
