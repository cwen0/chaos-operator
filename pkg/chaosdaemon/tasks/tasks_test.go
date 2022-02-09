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

package tasks

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaoserr"
)

type FakeConfig struct {
	i int
}

func (f *FakeConfig) Add(a Addable) error {
	A, OK := a.(*FakeConfig)
	if OK {
		f.i += A.i
		return nil
	}
	return errors.Wrapf(ErrCanNotAdd, "expect type : *FakeConfig, got : %T", a)
}

func (f *FakeConfig) Assign(c ChaosOnProcess) error {
	C, OK := c.(*FakeChaos)
	if OK {
		C.C.i = f.i
		return nil
	}
	return errors.Wrapf(ErrCanNotAssign, "expect type : *FakeChaos, got : %T", c)
}

func (f *FakeConfig) New(immutableValues interface{}) (ChaosOnProcess, error) {
	temp := immutableValues.(*FakeChaos)
	f.Assign(temp)
	return temp, nil
}

type FakeChaos struct {
	C              FakeConfig
	ErrWhenRecover bool
	ErrWhenInject  bool
	logger         logr.Logger
}

func (f *FakeChaos) Inject(pid PID) error {
	if f.ErrWhenInject {
		return chaoserr.NotImplemented("inject")
	}
	return nil
}

func (f *FakeChaos) Recover(pid PID) error {
	if f.ErrWhenRecover {
		return chaoserr.NotImplemented("recover")
	}
	return nil
}

func TestTasks(t *testing.T) {
	var log logr.Logger

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	log = zapr.NewLogger(zapLog)

	m := NewChaosOnProcessManager(log)

	chaos := FakeChaos{
		ErrWhenRecover: false,
		ErrWhenInject:  false,
		logger:         log,
	}
	task1 := FakeConfig{i: 1}
	uid1 := "1"
	err = m.Create(uid1, 1, &task1, &chaos)
	assert.NoError(t, err)
	err = m.Apply(uid1, 1, &task1)
	assert.Equal(t, errors.Cause(err), chaoserr.ErrDuplicateEntity)
	err = m.Recover(uid1, 1)
	assert.NoError(t, err)
	err = m.Recover(uid1, 1)
	assert.Equal(t, errors.Cause(err), chaoserr.NotFound("PID"))

	chaos.ErrWhenInject = true
	tasks2 := FakeConfig{i: 1}
	err = m.Create(uid1, 1, &tasks2, &chaos)
	assert.Equal(t, errors.Cause(err), chaoserr.NotImplemented("inject"))
	_, err = m.GetWithUID(uid1)
	assert.Equal(t, errors.Cause(err), chaoserr.NotFound("UID"))

	chaos.ErrWhenInject = false
	chaos.ErrWhenRecover = true
	tasks3 := FakeConfig{i: 1}
	err = m.Create(uid1, 1, &tasks3, &chaos)
	assert.NoError(t, err)
	err = m.Recover(uid1, 1)
	assert.Equal(t, errors.Cause(err), chaoserr.NotImplemented("recover"))
	p, err := m.GetWithPID(1)
	inner := p.(*FakeChaos)
	inner.ErrWhenRecover = false
	err = m.Recover(uid1, 1)
	assert.NoError(t, err)
}
