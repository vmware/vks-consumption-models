IMAGE ?= todo-app
TAG ?= latest
REGISTRY ?= quay.io/rahulk10
CHART_DIR ?= helm-charts
RELEASE ?= todo-app
NAMESPACE ?= default

.PHONY: docker-build docker-push helm-deploy helm-upgrade

docker-build:
	docker build -t $(REGISTRY)/$(IMAGE):$(TAG) .

docker-push: docker-build
	docker push $(REGISTRY)/$(IMAGE):$(TAG)

helm-deploy:
	helm upgrade --install $(RELEASE) $(CHART_DIR) \
	  --namespace $(NAMESPACE) --create-namespace \
	  --set image.repository=$(REGISTRY)/$(IMAGE) \
	  --set image.tag=$(TAG)

helm-upgrade: helm-deploy
