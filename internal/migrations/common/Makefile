GOPATH=$(shell go env GOPATH)
BINDATA = $(GOPATH)/bin/go-bindata

clear_bin_data:
	rm -f ./bindata.go

compress_migrations: clear_bin_data
ifeq (,$(wildcard ./glob.c))
	$(BINDATA) -ignore=Makefile -pkg=common ./...
else
	go get -u github.com/jteeuwen/go-bindata/...
	$(BINDATA) -ignore=Makefile -pkg=common ./...
endif