#!/usr/bin/env bash
# Copyright 2021 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

function must_contains() {
  message=$1
  substring=$2
  match=""
  if [ "$3" = "false" ]; then
      match="-v"
  fi

  echo $message | grep $match "$substring"
  if [ "$?" != "0" ]; then
      echo "'$substring' not found in '$message'"
      exit 1
  fi
}

echo "Deploy web-show for testing"
wget https://mirrors.chaos-mesh.org/v1.1.2/web-show/deploy.sh
bash deploy.sh

echo "Deploy busybox for test"
kubectl run busybox --image=radial/busyboxplus:curl -- sleep 3600

echo "wait pods status to running"
for ((k=0; k<30; k++)); do
    if [ $(kubectl get pods -l app=web-show | grep "Running" | wc -l) == 1 ] && [ $(kubectl get pods busybox | grep "Running" | wc -l) == 1 ]; then
        break
    fi
    sleep 1
done

echo "Confirm web-show works well"
must_contains "$(kubectl exec busybox -- sh -c "curl -I web-show.default:8081")" "HTTP/1.1 200 OK" true

echo "Run networkchaos"

echo "Run httpchaos"

cat <<EOF >delay.yaml
apiVersion: chaos-mesh.org/v1alpha1
kind: HTTPChaos
metadata:
  name: web-show-http-delay
spec:
  mode: one # the mode to run chaos action; supported modes are one/all/fixed/fixed-percent/random-max-percent
  selector: # pods where to inject chaos actions
    namespaces:
      - default
    labelSelectors:
      "app": "web-show"  # the label of the pod for chaos injection
  target: Response
  port: 8081
  path: "*"
  replace:
    code: 404
EOF
kubectl apply -f delay.yaml

echo "Confirm httpchaos works well"
sleep 5 # TODO: better way to wait for chaos being injected
must_contains "$(kubectl exec busybox -- sh -c "curl -I web-show.default:8081")" "HTTP/1.1 404 Not Found" true

echo "Recover"
./bin/chaosctl recover httpchaos -l app=web-show

echo "Confirm httpchaos recovered"
must_contains "$(kubectl exec busybox -- sh -c "curl -I web-show.default:8081")" "HTTP/1.1 200 OK" true

echo "Cleaning up httpchaos"
kubectl delete -f delay.yaml
rm delay.yaml

kubectl delete pod busybox
bash deploy.sh -d
rm deploy.sh
