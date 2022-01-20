package tasks

import (
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/ChaosErr"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
		return ChaosErr.NotImplemented("inject")
	}
	return nil
}

func (f *FakeChaos) Recover(pid PID) error {
	if f.ErrWhenRecover {
		return ChaosErr.NotImplemented("recover")
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
	assert.Equal(t, errors.Cause(err), ChaosErr.ErrDuplicateEntity)
	err = m.Recover(uid1, 1)
	assert.NoError(t, err)
	err = m.Recover(uid1, 1)
	assert.Equal(t, errors.Cause(err), ChaosErr.NotFound("PID"))

	chaos.ErrWhenInject = true
	tasks2 := FakeConfig{i: 1}
	err = m.Create(uid1, 1, &tasks2, &chaos)
	assert.Equal(t, errors.Cause(err), ChaosErr.NotImplemented("inject"))
	_, err = m.GetWithUID(uid1)
	assert.Equal(t, errors.Cause(err), ChaosErr.NotFound("UID"))

	chaos.ErrWhenInject = false
	chaos.ErrWhenRecover = true
	tasks3 := FakeConfig{i: 1}
	err = m.Create(uid1, 1, &tasks3, &chaos)
	assert.NoError(t, err)
	err = m.Recover(uid1, 1)
	assert.Equal(t, errors.Cause(err), ChaosErr.NotImplemented("recover"))
	p, err := m.GetWithPID(1)
	inner := p.(*FakeChaos)
	inner.ErrWhenRecover = false
	err = m.Recover(uid1, 1)
	assert.NoError(t, err)
}
