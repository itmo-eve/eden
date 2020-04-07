# HOSTARCH is the host architecture
# ARCH is the target architecture
# we need to keep track of them separately
HOSTARCH ?= $(shell uname -m)
HOSTOS ?= $(shell uname -s | tr A-Z a-z)

# canonicalized names for host architecture
override HOSTARCH := $(subst aarch64,arm64,$(subst x86_64,amd64,$(HOSTARCH)))

# unless otherwise set, I am building for my own architecture, i.e. not cross-compiling
# and for my OS
ARCH ?= $(HOSTARCH)
OS ?= $(HOSTOS)

# canonicalized names for target architecture
override ARCH := $(subst aarch64,arm64,$(subst x86_64,amd64,$(ARCH)))

DIST=$(CURDIR)/dist
BINDIR := $(DIST)/bin
BIN := eden
LOCALBIN := $(BINDIR)/$(BIN)-$(OS)-$(ARCH)
CONFIG=$(DIST)/Config.in
include $(CONFIG)


ZARCH ?= $(HOSTARCH)
export ZARCH

clean: stop
	rm -rf $(DIST)/* || sudo rm -rf $(DIST)/*

$(DIST):
	mkdir -p $@
$(BINDIR):
	mkdir -p $@

EVE_DIST=$(DIST)/eve

BIOS_IMG ?= $(EVE_DIST)/dist/$(ZARCH)/OVMF.fd
LIVE_IMG ?= $(EVE_DIST)/dist/$(ZARCH)/live.img
EVE_URL ?= "https://github.com/lf-edge/eve.git"
EVE_REF ?= "master"
ADAM_URL ?= "https://github.com/lf-edge/adam.git"
ADAM_REF ?= "master"

EVE_REF_OLD=$(EVE_REF)
EVE_DIST_OLD=$(EVE_REF)

EVE_BASE_REF?=4.10.0
EVE_BASE_VERSION?=$(EVE_BASE_REF)-$(ZARCH)
EVE_BASE_DIST=$(DIST)/evebaseos

$(EVE_DIST):
ifneq ($(EVE_REF),)
	git clone --branch $(EVE_REF) --single-branch $(EVE_URL) $(EVE_DIST)
else
	git clone $(EVE_URL) $(EVE_DIST)
endif

ADAM_DIST ?= $(DIST)/adam

$(ADAM_DIST):
ifneq ($(ADAM_REF),)
	git clone --branch $(ADAM_REF) --single-branch $(ADAM_URL) $(ADAM_DIST)
else
	git clone $(ADAM_URL) $(ADAM_DIST)
endif

IMAGE_DIST ?= $(DIST)/images

$(IMAGE_DIST):
	mkdir -p $@

# any non-empty value will trigger eve rebuild
REBUILD ?=

# any non-empty value will trigger fix eve ip
FIX_IP ?=

ACCEL ?=
SSH_PORT ?= 2222

run: eserver_run adam_run eve_run

$(CONFIG): save

save: $(DIST)
	@echo "# Configuration settings" > $(CONFIG)
	@echo ADAM_DIST=$(ADAM_DIST) >> $(CONFIG)
	@echo ZARCH=$(ZARCH) >> $(CONFIG)
	@echo BIOS_IMG=$(BIOS_IMG) >> $(CONFIG)
	@echo LIVE_IMG=$(LIVE_IMG) >> $(CONFIG)
	@echo EVE_URL=$(EVE_URL) >> $(CONFIG)
	@echo EVE_REF=$(EVE_REF) >> $(CONFIG)
	@echo ADAM_URL=$(ADAM_URL) >> $(CONFIG)
	@echo ADAM_REF=$(ADAM_REF) >> $(CONFIG)
	@echo ACCEL=$(ACCEL) >> $(CONFIG)
	@echo SSH_PORT=$(SSH_PORT) >> $(CONFIG)
	@echo CERTS_DIST=$(CERTS_DIST) >> $(CONFIG)
	@echo DOMAIN=$(DOMAIN) >> $(CONFIG)
	@echo IP=$(IP) >> $(CONFIG)
	@echo UUID=$(UUID) >> $(CONFIG)
	@echo ADAM_PORT=$(ADAM_PORT) >> $(CONFIG)
	@echo EVE_BASE_REF=$(EVE_BASE_REF) >> $(CONFIG)
	@echo EVE_BASE_VERSION=$(EVE_BASE_VERSION) >> $(CONFIG)
	@echo ADAM_CA=$(ADAM_DIST)/run/config/root-certificate.pem >> $(CONFIG)
	@echo EVE_CERT=$(ADAM_DIST)/run/config/onboard.cert.pem >> $(CONFIG)

eve_run: save eve_stop eve_live
	@echo EVE run
	nohup make -C $(EVE_DIST) CONF_DIR=$(ADAM_DIST)/run/config/ SSH_PORT=$(SSH_PORT) ACCEL=$(ACCEL) run >$(DIST)/eve.log 2>&1 & echo "$$!" >$(DIST)/eve.pid
	@echo You can see logs of EVE:
	@echo $(DIST)/eve.log
	@echo You can ssh into EVE:
	@echo ssh -i $(CERTS_DIST)/id_rsa -p $(SSH_PORT) root@127.0.0.1

IMGS := $(LIVE_IMG) $(BIOS_IMG)
IMGS_MISSING := $(shell $(foreach f,$(IMGS),test -e "$f" || echo "$f";))

eve_rootfs: $(EVE_DIST)
	make -C $(EVE_DIST) CONF_DIR=$(ADAM_DIST)/run/config/ rootfs

eve_live: $(EVE_DIST)
ifneq ($(FIX_IP),)
	chmod a+x $(CURDIR)/scripts/fixIPs.sh
	$(CURDIR)/scripts/fixIPs.sh $(EVE_DIST)/Makefile
endif
ifneq ($(REBUILD),)
	make -C $(EVE_DIST) CONF_DIR=$(ADAM_DIST)/run/config/ live
else
ifneq ($(IMGS_MISSING),)
		make -C $(EVE_DIST) CONF_DIR=$(ADAM_DIST)/run/config/ live
else
		true
endif
endif


CERTS_DIST ?= $(DIST)/certs
DOMAIN ?= mydomain.adam

ifeq ($(shell uname -s), Darwin)
IP?=$(shell ifconfig|grep "inet "|grep -Fv 127.0.0.1|awk '{print $$2}'|tail -1)
else
IP?=$(shell hostname -I | cut -d' ' -f1)
endif
UUID?=$(shell uuidgen)
ADAM_PORT?=3333

adam_docker_stop:
	docker ps|grep eden_adam&&docker stop eden_adam||echo ""
	docker ps --all|grep eden_adam&&docker rm eden_adam||echo ""

adam_run: save adam_docker_stop $(ADAM_DIST) certs_and_config
	@echo Adam Run
	cd $(ADAM_DIST); docker run --name eden_adam -d -v $(ADAM_DIST)/run:/adam/run -p $(ADAM_PORT):8080 lfedge/adam server --conf-dir /tmp
	@echo ADAM is ready for access: https://$(IP):$(ADAM_PORT)

$(CERTS_DIST):
	test -d $@ || mkdir -p $@

certs_and_config: $(CERTS_DIST) ecerts
ifeq ("$(wildcard $(ADAM_DIST)/run/adam/server.pem)","")
	test -d $(ADAM_DIST)/run/adam || mkdir -p $(ADAM_DIST)/run/adam
	test -d $(ADAM_DIST)/run/config || mkdir -p $(ADAM_DIST)/run/config
	$(LOCALBIN) certs -o $(CERTS_DIST) -i $(IP) -d $(DOMAIN) -u $(UUID)
	cp $(CERTS_DIST)/root-certificate.pem $(ADAM_DIST)/run/config/
	cp $(CERTS_DIST)/onboard.cert.pem $(ADAM_DIST)/run/config/
	cp $(CERTS_DIST)/onboard.key.pem $(ADAM_DIST)/run/config/
	cp $(CERTS_DIST)/server.pem $(ADAM_DIST)/run/adam/
	cp $(CERTS_DIST)/server-key.pem $(ADAM_DIST)/run/adam/
	echo $(IP) $(DOMAIN) >$(ADAM_DIST)/run/config/hosts
	echo $(DOMAIN):$(ADAM_PORT) >$(ADAM_DIST)/run/config/server
	test -f $(CERTS_DIST)/id_rsa.pub || ssh-keygen -t rsa -f $(CERTS_DIST)/id_rsa -q -N ""
	yes | cp -f $(CERTS_DIST)/id_rsa.pub $(ADAM_DIST)/run/config/authorized_keys
endif

stop: adam_docker_stop eve_stop eserver_stop

eve_stop:
	test -f $(DIST)/eve.pid && kill $(shell cat $(DIST)/eve.pid) && rm $(DIST)/eve.pid || echo ""

test_controller:
	ADAM_IP=$(IP) ADAM_DIST=$(ADAM_DIST) EVE_BASE_REF=$(EVE_BASE_REF) ZARCH=$(ZARCH) ADAM_PORT=$(ADAM_PORT) ADAM_CA=$(ADAM_CA) EVE_CERT=$(EVE_CERT) go test ./tests/integration/controller_test.go ./tests/integration/common.go -v -count=1 -timeout 3000s

test_base_image: test_controller
	ADAM_IP=$(IP) ADAM_DIST=$(ADAM_DIST) EVE_BASE_REF=$(EVE_BASE_REF) ZARCH=$(ZARCH) ADAM_PORT=$(ADAM_PORT) ADAM_CA=$(ADAM_CA) EVE_CERT=$(EVE_CERT) go test ./tests/integration/baseImage_test.go ./tests/integration/common.go -v -count=1 -timeout 4500s

test_network_instance: test_controller
	ADAM_IP=$(IP) ADAM_DIST=$(ADAM_DIST) EVE_BASE_REF=$(EVE_BASE_REF) ZARCH=$(ZARCH) ADAM_PORT=$(ADAM_PORT) ADAM_CA=$(ADAM_CA) EVE_CERT=$(EVE_CERT) go test ./tests/integration/networkInstance_test.go ./tests/integration/common.go -v -count=1 -timeout 4000s

test: test_base_image test_network_instance

bin: $(LOCALBIN)
build: $(LOCALBIN) $(BIN)
$(LOCALBIN): $(BINDIR)
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o $@ .
$(BIN):
	@if [ "$(OS)" = "$(HOSTOS)" -a "$(ARCH)" = "$(HOSTARCH)" -a ! -e "$@" ]; then ln -s $(LOCALBIN) $@; fi

SHA256_CMD = sha256sum
ifeq ($(shell uname -s), Darwin)
        SHA256_CMD = openssl sha256 -r
endif

$(IMAGE_DIST)/baseos.qcow2: save $(IMAGE_DIST) certs_and_config
ifeq ("$(wildcard $(IMAGE_DIST)/baseos.qcow2)","")
	$(MAKE) eve_rootfs EVE_REF=$(EVE_BASE_REF) EVE_DIST=$(EVE_BASE_DIST)
	cp $(EVE_BASE_DIST)/dist/$(ZARCH)/installer/rootfs.img $(IMAGE_DIST)/baseos.qcow2
	cd $(IMAGE_DIST); $(SHA256_CMD) baseos.qcow2>baseos.sha256
	echo EVE_VERSION>$(IMAGE_DIST)/version.yml.in
	$(MAKE) -C $(EVE_BASE_DIST) $(IMAGE_DIST)/version.yml
	$(MAKE) save_ref_dist_base EVE_REF=$(EVE_REF_OLD) EVE_DIST=$(EVE_DIST_OLD)
	rm -rf $(IMAGE_DIST)/version.yml.in
	rm -rf $(IMAGE_DIST)/version.yml
endif

save_ref_dist_base:
	$(eval EVE_BASE_VERSION := $(shell cat $(IMAGE_DIST)/version.yml))
	$(MAKE) save EVE_REF=$(EVE_REF_OLD) EVE_DIST=$(EVE_DIST_OLD) EVE_BASE_VERSION=$(EVE_BASE_VERSION)

ESERVER_PORT=8888

eserver_run: build $(IMAGE_DIST)/baseos.qcow2 eserver_stop
	@echo eden server run
	nohup $(LOCALBIN) server -p $(ESERVER_PORT) -d $(IMAGE_DIST) 2>&1 >/dev/null & echo "$$!" >$(DIST)/eserver.pid
	@echo eden server run on port $(ESERVER_PORT)

eserver_stop:
	test -f $(DIST)/eserver.pid && kill $(shell cat $(DIST)/eserver.pid) && rm $(DIST)/eserver.pid || echo ""

help:
	@echo "EDEN is the harness for testing EVE and ADAM"
	@echo
	@echo "This Makefile automates commons tasks of building and running"
	@echo "  * EVE"
	@echo "  * ADAM"
	@echo
	@echo "Commonly used maintenance and development targets:"
	@echo "   run           run ADAM and EVE"
	@echo "   test          run tests"
	@echo "   stop          stop ADAM and EVE"
	@echo "   clean         cleanup directories"
	@echo "   bin		build utilities"
	@echo
	@echo "You need access to docker socket and installed qemu packages"
	@echo "Use of ACCEL=true is recommended"
	@echo "You must set the FIX_IP=true variable if you use subnets 192.168.1.0/24 or 192.168.2.0/24 for any interface on host"
