SHELL=/bin/bash

parallel=false

.ONESHELL:

.PHONY: testacc

testacc:
ifeq ($(parallel),true)
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
else
	TF_ACC=1 go test -p 1 ./... -v $(TESTARGS) -timeout 120m
endif

copydocs:
	$(shell ./copydocs.sh)