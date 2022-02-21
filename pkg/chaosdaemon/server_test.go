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
	stdlog "log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	logutil "github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	rootLogger, err := logutil.NewDefaultZapLogger()
	if err != nil {
		stdlog.Fatal("failed to create root logger", err)
	}
	Context("newDaemonServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			_, err := newDaemonServer(crclients.ContainerRuntimeContainerd, nil, rootLogger)
			Expect(err).To(BeNil())
		})

		It("should fail on CreateContainerRuntimeInfoClient", func() {
			_, err := newDaemonServer("invalid-runtime", nil, rootLogger)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("newGRPCServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &test.MockClient{})()
			daemonServer, err := newDaemonServer(crclients.ContainerRuntimeContainerd, nil, rootLogger)
			Expect(err).To(BeNil())
			_, err = newGRPCServer(daemonServer, &MockRegisterer{}, tlsConfig{})
			Expect(err).To(BeNil())
		})

		It("should panic", func() {
			Ω(func() {
				defer mock.With("MockContainerdClient", &test.MockClient{})()
				defer mock.With("PanicOnMustRegister", "mock panic")()
				daemonServer, err := newDaemonServer(crclients.ContainerRuntimeContainerd, nil, rootLogger)
				Expect(err).To(BeNil())
				_, err = newGRPCServer(daemonServer, &MockRegisterer{}, tlsConfig{})
				Expect(err).To(BeNil())
			}).Should(Panic())
		})
	})
})

type MockRegisterer struct {
	RegisterGatherer
}

func (*MockRegisterer) MustRegister(...prometheus.Collector) {
	if err := mock.On("PanicOnMustRegister"); err != nil {
		panic(err)
	}
}
