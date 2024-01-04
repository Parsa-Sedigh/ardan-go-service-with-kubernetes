# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH := /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# ==============================================================================
# CLASS NOTES
#
# Kind
# 	For full Kind v0.20 release notes: https://github.com/kubernetes-sigs/kind/releases/tag/v0.20.0

# RSA Keys
# 	To generate a private/public key PEM file(right from the CLI).
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 	$ openssl rsa -pubout -in private.pem -out public.pem
#
# OPA Playground
# 	https://play.openpolicyagent.org/
# 	https://academy.styra.com/
# 	https://www.openpolicyagent.org/docs/latest/policy-reference/

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.21.4
ALPINE          := alpine:3.18
KIND            := kindest/node:v1.27.3

# We're specifying the major.minor version and not the patch. But in a prod env, you wanna specify the patch@<SHA> as well.
POSTGRES        := postgres:15.5
VAULT           := hashicorp/vault:1.15
GRAFANA         := grafana/grafana:10.2.0
PROMETHEUS      := prom/prometheus:v2.48.0
TEMPO           := grafana/tempo:2.3.0
LOKI            := grafana/loki:2.9.0
PROMTAIL        := grafana/promtail:2.9.0
TELEPRESENCE    := datawire/tel2:2.13.1

# name of the cluster
KIND_CLUSTER    := ardan-starter-cluster
NAMESPACE       := sales-system
APP             := sales
BASE_IMAGE_NAME := ardanlabs/service
SERVICE_NAME    := sales-api
VERSION         := 0.0.5
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)

# VERSION       := "0.0.1-$(shell git rev-parse --short HEAD)"

# ==============================================================================
# Building containers

all: service

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Running from within k8s/kind

dev-bill:
	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-up-local:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/dev/kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)

	# load telepresence image into our kind envionrment.
	#kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)

dev-up: dev-up-local
	# runs a helm chart provided by the telepresence tooling that is responsible for starting the telepresence service inside the cluster
	# when this line runs, there's gonna be a new namespace named `ambassador` and the a new pod called traffic-manager-<random string> which is
	# the telepresence pod inside the cluster
	telepresence --context=kind-$(KIND_CLUSTER) helm install
	# responsible for starting the agent service behind the scenes(outside of the cluster) and create that tunnel for us
	telepresence --context=kind-$(KIND_CLUSTER) connect

dev-down-local:
	kind delete cluster --name $(KIND_CLUSTER)

dev-down:
	telepresence quit -s
	kind delete cluster --name $(KIND_CLUSTER)

dev-load:
	# kind can go to docker registry, but if we preload all the images, we don't need to go out to docker registry
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER);

dev-apply:
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	# kustomize build is gonna generate all the yaml and pipe the result to kubectl apply in order to apply them to k8s.
	# Then we wait for the condition=Ready to know that the service is up and running
	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
    kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --for=condition=Ready

# ==============================================================================

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

# in our development environment, this is how we're gonna restart the pod when we update the image. You wouldn't do this in
# staging or production environment, there's better ways to upload the image. Normally it has a new tag, the environment sees that
# and it starts a canary or rolling update. But we don't any of that in development env.
dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

# 1. builds everything
# 2. loads them in kind environment
# 3. restart the pods
# So anytime we're updating the code, we're gonna run `dev-update`
dev-update: all dev-load dev-restart

# Note: Most of the time when you're changing the configuration, k8s is gonna auto-restart the pod for you. So we don't need to do the
# `dev-restart`, we get the restart automagically. So we're gonna decide did we change configuration(k8s config) or we just
# changed the binary(application binary). If we didn't change config, you don't get a restart.
dev-update-apply: all dev-load dev-apply

# ==============================================================================

dev-logs:
	# tail=100 means start with the last 100 logs
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

dev-describe-deployment:
	kubectl describe deployment --namespace=$(NAMESPACE) $(APP)

dev-describe-sales:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(APP)

# ==============================================================================

run-scratch:
	go run app/tooling/scratch/main.go

run-local:
	go run app/services/sales-api/main.go

run-local-help:
	go run app/services/sales-api/main.go --help

tidy:
	go mod tidy
	go mod vendor

metrics-view:
	expvarmon -ports="$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-local:
	expvarmon -ports="localhost:4000" -endpoint="/metrics" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

test-endpoint:
	curl -il $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:3000/test

test-endpoint-local:
	curl -il localhost:3000/test

test-endpoint-auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:3000/test/auth

test-endpoint-auth-local:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/test/auth

liveness-local:
	curl -il http://localhost:4000/debug/liveness

liveness:
	curl -il $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000/debug/liveness

readiness-local:
	curl -il http://localhost:4000/debug/readiness

readiness:
	curl -il $(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000/debug/readiness

pgcli-local:
	pgcli postgresql://postgres:postgres@localhost

pgcli:
	pgcli postgresql://postgres:postgres@database-service.$(NAMESPACE).svc.cluster.local