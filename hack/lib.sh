#!/usr/bin/env bash

# Copyright 2020 PingCAP, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# See the License for the specific language governing permissions and
# limitations under the License.

if [ -z "$ROOT" ]; then
    echo "error: ROOT should be initialized"
    exit 1
fi

OS=$(go env GOOS)
ARCH=$(go env GOARCH)
OUTPUT=${ROOT}/output
OUTPUT_BIN=${OUTPUT}/bin
TERRAFORM_BIN=${OUTPUT_BIN}/terraform
TERRAFORM_VERSION=0.12.12
KUBECTL_VERSION=1.12.10
KUBECTL_BIN=$OUTPUT_BIN/kubectl
HELM_BIN=$OUTPUT_BIN/helm
#
# Don't ugprade to 2.15.x/2.16.x until this issue
# (https://github.com/helm/helm/issues/6361) has been fixed.
#
HELM_VERSION=2.9.1
KIND_VERSION=0.7.0
KIND_BIN=$OUTPUT_BIN/kind
KUBEBUILDER_BIN=$OUTPUT_BIN/kubebuilder
KUBEBUILDER_VERSION=2.2.0
KUSTOMIZE_BIN=$OUTPUT_BIN/kustomize
KUSTOMIZE_VERSION=3.5.4

test -d "$OUTPUT_BIN" || mkdir -p "$OUTPUT_BIN"

function hack::verify_terraform() {
    if test -x "$TERRAFORM_BIN"; then
        local v=$($TERRAFORM_BIN version | awk '/^Terraform v.*/ { print $2 }' | sed 's#^v##')
        [[ "$v" == "$TERRAFORM_VERSION" ]]
        return
    fi
    return 1
}

function hack::install_terraform() {
    echo "Installing terraform v$TERRAFORM_VERSION..."
    local tmpdir=$(mktemp -d)
    trap "test -d $tmpdir && rm -r $tmpdir" RETURN
    pushd $tmpdir > /dev/null
    wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${OS}_${ARCH}.zip
    unzip terraform_${TERRAFORM_VERSION}_${OS}_${ARCH}.zip
    mv terraform $TERRAFORM_BIN
    popd > /dev/null
    chmod +x $TERRAFORM_BIN
}

function hack::ensure_terraform() {
    if ! hack::verify_terraform; then
        hack::install_terraform
    fi
}

function hack::verify_kubectl() {
    if test -x "$KUBECTL_BIN"; then
        [[ "$($KUBECTL_BIN version --client --short | grep -o -E '[0-9]+\.[0-9]+\.[0-9]+')" == "$KUBECTL_VERSION" ]]
        return
    fi
    return 1
}

function hack::ensure_kubectl() {
    if hack::verify_kubectl; then
        return 0
    fi
    echo "Installing kubectl v$KUBECTL_VERSION..."
    tmpfile=$(mktemp)
    trap "test -f $tmpfile && rm $tmpfile" RETURN
    curl --retry 10 -L -o $tmpfile https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl
    mv $tmpfile $KUBECTL_BIN
    chmod +x $KUBECTL_BIN
}

function hack::verify_helm() {
    if test -x "$HELM_BIN"; then
        local v=$($HELM_BIN version --short --client | grep -o -E '[0-9]+\.[0-9]+\.[0-9]+')
        [[ "$v" == "$HELM_VERSION" ]]
        return
    fi
    return 1
}

function hack::ensure_helm() {
    if hack::verify_helm; then
        return 0
    fi
    local HELM_URL=http://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-${OS}-${ARCH}.tar.gz
    curl --retry 10 -L -s "$HELM_URL" | tar --strip-components 1 -C $OUTPUT_BIN -zxf - ${OS}-${ARCH}/helm
}

function hack::verify_kind() {
    if test -x "$KIND_BIN"; then
        [[ "$($KIND_BIN --version 2>&1 | cut -d ' ' -f 3)" == "$KIND_VERSION" ]]
        return
    fi
    return 1
}

function hack::ensure_kind() {
    if hack::verify_kind; then
        return 0
    fi
    tmpfile=$(mktemp)
    trap "test -f $tmpfile && rm $tmpfile" RETURN
    curl --retry 10 -L -o $tmpfile https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-$(uname)-amd64
    mv $tmpfile $KIND_BIN
    chmod +x $KIND_BIN
}

function hack::verify_kubebuilder() {
    if test -x "$KUBEBUILDER_BIN"; then
        v=$($KUBEBUILDER_BIN version | grep -o -E '[0-9]+\.[0-9]+\.[0-9]+' | head -n 1)
        [[ "${v}" == "${KUBEBUILDER_VERSION}" ]]
        return
    fi
    return 1
}

function hack::ensure_kubebuilder() {
    if hack::verify_kubebuilder; then
        return 0
    fi
    tmpfile=$(mktemp)
    trap "test -f $tmpfile && rm $tmpfile" RETURN
    curl --retry 10 -L -o ${tmpfile} https://go.kubebuilder.io/dl/$KUBEBUILDER_VERSION/$OS/$ARCH
    tar -C ${OUTPUT_BIN} -xf ${tmpfile} kubebuilder_${KUBEBUILDER_VERSION}_${OS}_${ARCH}/bin/kubebuilder
    mv ${OUTPUT_BIN}/kubebuilder_${KUBEBUILDER_VERSION}_${OS}_${ARCH}/bin/kubebuilder ${KUBEBUILDER_BIN}
    chmod +x ${KUBEBUILDER_BIN}
    rm -r ${OUTPUT_BIN}/kubebuilder_${KUBEBUILDER_VERSION}_${OS}_${ARCH}
}

function hack::verify_kustomize() {
    if test -x "$KUSTOMIZE_BIN"; then
        v=$($KUSTOMIZE_BIN version | grep -o -E '[0-9]+\.[0-9]+\.[0-9]+')
        [[ "${v}" == "${KUSTOMIZE_VERSION}" ]]
        return
    fi
    return 1
}

function hack::ensure_kustomize() {
    if hack::verify_kustomize; then
        return 0
    fi
    tmpfile=$(mktemp)
    trap "test -f $tmpfile && rm $tmpfile" RETURN
    curl --retry 10 -L -o ${tmpfile} "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_${OS}_${ARCH}.tar.gz"
    tar -C $OUTPUT_BIN -zxf ${tmpfile}
    chmod +x $KUSTOMIZE_BIN
}

# hack::version_ge "$v1" "$v2" checks whether "v1" is greater or equal to "v2"
function hack::version_ge() {
    [ "$(printf '%s\n' "$1" "$2" | sort -V | head -n1)" = "$2" ]
}
