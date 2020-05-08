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

package event

import (
	"context"

	"github.com/pingcap/chaos-mesh/pkg/core"
	"github.com/pingcap/chaos-mesh/pkg/store/dbstore"
)

// NewStore return a new EventStore.
func NewStore(db *dbstore.DB) core.EventStore {
	db.AutoMigrate(&core.Event{})

	return &eventStore{db}
}

type eventStore struct {
	db *dbstore.DB
}

// TODO: implement core.EventStore interface
func (e *eventStore) List(context.Context) ([]*core.Event, error) { return nil, nil }
func (e *eventStore) ListByExperiment(context.Context, string, string) ([]*core.Event, error) {
	return nil, nil
}
func (e *eventStore) Find(context.Context, int64) (*core.Event, error) { return nil, nil }
func (e *eventStore) Create(context.Context, *core.Event) error        { return nil }
func (e *eventStore) Update(context.Context, *core.Event) error        { return nil }
