# Generated by ./cmd/generate-makefile. DO NOT EDIT.

.PHONY: image
image: image-chaos-daemon image-chaos-mesh image-chaos-dashboard image-build-env image-dev-env image-e2e-helper image-chaos-mesh-e2e image-chaos-kernel image-chaos-jvm image-chaos-dlv ## Build all container images

.PHONY: image-chaos-daemon
image-chaos-daemon:images/chaos-daemon/.dockerbuilt ## Build container image for chaos-daemon, ghcr.io/chaos-mesh/chaos-daemon:latest

images/chaos-daemon/.dockerbuilt: SHELL=bash
images/chaos-daemon/.dockerbuilt: images/chaos-daemon/bin/chaos-daemon images/chaos-daemon/bin/pause images/chaos-daemon/bin/cdh images/chaos-daemon/Dockerfile
	$(ROOT)/build/build_image.py chaos-daemon images/chaos-daemon
	touch images/chaos-daemon/.dockerbuilt

.PHONY: image-chaos-mesh
image-chaos-mesh:images/chaos-mesh/.dockerbuilt ## Build container image for chaos-mesh, ghcr.io/chaos-mesh/chaos-mesh:latest

images/chaos-mesh/.dockerbuilt: SHELL=bash
images/chaos-mesh/.dockerbuilt: images/chaos-mesh/bin/chaos-controller-manager images/chaos-mesh/Dockerfile
	$(ROOT)/build/build_image.py chaos-mesh images/chaos-mesh
	touch images/chaos-mesh/.dockerbuilt

.PHONY: image-chaos-dashboard
image-chaos-dashboard:images/chaos-dashboard/.dockerbuilt ## Build container image for chaos-dashboard, ghcr.io/chaos-mesh/chaos-dashboard:latest

images/chaos-dashboard/.dockerbuilt: SHELL=bash
images/chaos-dashboard/.dockerbuilt: images/chaos-dashboard/bin/chaos-dashboard images/chaos-dashboard/Dockerfile
	$(ROOT)/build/build_image.py chaos-dashboard images/chaos-dashboard
	touch images/chaos-dashboard/.dockerbuilt

.PHONY: image-build-env
image-build-env:images/build-env/.dockerbuilt ## Build container image for build-env, ghcr.io/chaos-mesh/build-env:latest

images/build-env/.dockerbuilt: SHELL=bash
images/build-env/.dockerbuilt:  images/build-env/Dockerfile
	$(ROOT)/build/build_image.py build-env images/build-env
	touch images/build-env/.dockerbuilt

.PHONY: image-dev-env
image-dev-env:images/dev-env/.dockerbuilt ## Build container image for build-env, ghcr.io/chaos-mesh/dev-env:latest

images/dev-env/.dockerbuilt: SHELL=bash
images/dev-env/.dockerbuilt:  images/dev-env/Dockerfile
	$(ROOT)/build/build_image.py dev-env images/dev-env
	touch images/dev-env/.dockerbuilt

.PHONY: image-e2e-helper
image-e2e-helper:e2e-test/cmd/e2e_helper/.dockerbuilt ## Build container image for e2e-helper

e2e-test/cmd/e2e_helper/.dockerbuilt: SHELL=bash
e2e-test/cmd/e2e_helper/.dockerbuilt:  e2e-test/cmd/e2e_helper/Dockerfile
	$(ROOT)/build/build_image.py e2e-helper e2e-test/cmd/e2e_helper
	touch e2e-test/cmd/e2e_helper/.dockerbuilt

.PHONY: image-chaos-mesh-e2e
image-chaos-mesh-e2e:e2e-test/image/e2e/.dockerbuilt ## Build container image for running e2e tests

e2e-test/image/e2e/.dockerbuilt: SHELL=bash
e2e-test/image/e2e/.dockerbuilt: e2e-test/image/e2e/manifests e2e-test/image/e2e/chaos-mesh e2e-build e2e-test/image/e2e/Dockerfile
	$(ROOT)/build/build_image.py chaos-mesh-e2e e2e-test/image/e2e
	touch e2e-test/image/e2e/.dockerbuilt

.PHONY: image-chaos-kernel
image-chaos-kernel:images/chaos-kernel/.dockerbuilt ## Build container image for chaos-kernel, ghcr.io/chaos-mesh/chaos-kernel:latest

images/chaos-kernel/.dockerbuilt: SHELL=bash
images/chaos-kernel/.dockerbuilt:  images/chaos-kernel/Dockerfile
	$(ROOT)/build/build_image.py chaos-kernel images/chaos-kernel
	touch images/chaos-kernel/.dockerbuilt

.PHONY: image-chaos-jvm
image-chaos-jvm:images/chaos-jvm/.dockerbuilt ## (Deprecated) Build container image for chaos-jvm

images/chaos-jvm/.dockerbuilt: SHELL=bash
images/chaos-jvm/.dockerbuilt:  images/chaos-jvm/Dockerfile
	$(ROOT)/build/build_image.py chaos-jvm images/chaos-jvm
	touch images/chaos-jvm/.dockerbuilt

.PHONY: image-chaos-dlv
image-chaos-dlv:images/chaos-dlv/.dockerbuilt ## Build container image for chaos-dlv

images/chaos-dlv/.dockerbuilt: SHELL=bash
images/chaos-dlv/.dockerbuilt:  images/chaos-dlv/Dockerfile
	$(ROOT)/build/build_image.py chaos-dlv images/chaos-dlv
	touch images/chaos-dlv/.dockerbuilt

.PHONY: clean-image-built
clean-image-built:
	rm -f images/chaos-daemon/.dockerbuilt
	rm -f images/chaos-mesh/.dockerbuilt
	rm -f images/chaos-dashboard/.dockerbuilt
	rm -f images/build-env/.dockerbuilt
	rm -f images/dev-env/.dockerbuilt
	rm -f e2e-test/cmd/e2e_helper/.dockerbuilt
	rm -f e2e-test/image/e2e/.dockerbuilt
	rm -f images/chaos-kernel/.dockerbuilt
	rm -f images/chaos-jvm/.dockerbuilt
	rm -f images/chaos-dlv/.dockerbuilt

