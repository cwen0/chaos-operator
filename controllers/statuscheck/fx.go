// Copyright Chaos Mesh Authors.
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

package statuscheck

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/go-logr/logr"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func Bootstrap(mgr ctrl.Manager, client client.Client, logger logr.Logger, recorderBuilder *recorder.RecorderBuilder) error {
	if !config.ShouldSpawnController("statuscheck") {
		return nil
	}

	return builder.Default(mgr).
		For(&v1alpha1.StatusCheck{}).
		Named("statuscheck").
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObj := e.ObjectOld.(*v1alpha1.StatusCheck)
				newObj := e.ObjectNew.(*v1alpha1.StatusCheck)

				return !reflect.DeepEqual(oldObj.Spec, newObj.Spec)
			},
		}).
		Complete(NewReconciler(client, logger.WithName("statuscheck"), recorderBuilder.Build("statuscheck")))
}
