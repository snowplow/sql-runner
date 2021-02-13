.PHONY: all format lint tidy setup-reset setup-up setup-down test goveralls integration release release-dry clean

# -----------------------------------------------------------------------------
#  CONSTANTS
# -----------------------------------------------------------------------------

src_dir = sql_runner

build_dir = build

coverage_dir  = $(build_dir)/coverage
coverage_out  = $(coverage_dir)/coverage.out
coverage_html = $(coverage_dir)/coverage.html

output_dir    = $(build_dir)/output

linux_dir     = $(output_dir)/linux
darwin_dir    = $(output_dir)/darwin
windows_dir   = $(output_dir)/windows

bin_name      = sql-runner
bin_linux     = $(linux_dir)/$(bin_name)
bin_darwin    = $(darwin_dir)/$(bin_name)
bin_windows   = $(windows_dir)/$(bin_name)

version       = $(shell cat VERSION)

artifacts_dir = $(build_dir)/dist

zip_suffix    = amd64.zip
zip_prefix    = sql_runner_$(subst -,_,$(version))

zip_linux     = $(zip_prefix)_linux_$(zip_suffix)
zip_darwin    = $(zip_prefix)_darwin_$(zip_suffix)
zip_windows   = $(zip_prefix)_windows_$(zip_suffix)


# -----------------------------------------------------------------------------
#  BUILDING
# -----------------------------------------------------------------------------

all:
	GO111MODULE=on go get -u github.com/mitchellh/gox
	GO111MODULE=on CGO_ENABLED=0 gox -osarch=linux/amd64 -output=$(bin_linux) ./$(src_dir)
	GO111MODULE=on CGO_ENABLED=0 gox -osarch=darwin/amd64 -output=$(bin_darwin) ./$(src_dir)
	GO111MODULE=on CGO_ENABLED=0 gox -osarch=windows/amd64 -output=$(bin_windows) ./$(src_dir)

# -----------------------------------------------------------------------------
#  FORMATTING
# -----------------------------------------------------------------------------

format:
	GO111MODULE=on go fmt ./$(src_dir)
	GO111MODULE=on gofmt -s -w ./$(src_dir)

lint:
	GO111MODULE=on go get -u golang.org/x/lint/golint
	GO111MODULE=on golint ./$(src_dir)

tidy:
	GO111MODULE=on go mod tidy

# -----------------------------------------------------------------------------
#  TESTING
# -----------------------------------------------------------------------------

setup-reset: setup-down setup-up

setup-up:
	docker-compose -f ./integration/docker-compose.yml up -d
	sleep 2
	./integration/setup_consul.sh

setup-down:
	docker-compose -f ./integration/docker-compose.yml down

test:
	mkdir -p $(coverage_dir)
	GO111MODULE=on go get -u golang.org/x/tools/cmd/cover
	GO111MODULE=on go test ./$(src_dir) -tags test -v -covermode=count -coverprofile=$(coverage_out)
	GO111MODULE=on go tool cover -html=$(coverage_out) -o $(coverage_html)

goveralls: test
	GO111MODULE=on go get -u github.com/mattn/goveralls
	goveralls -coverprofile=$(coverage_out) -service=github

integration:
ifndef DISTRO
	$(error DISTRO is undefined - this should be set to 'linux' or 'darwin'!)
endif
	./integration/run_tests.sh

# -----------------------------------------------------------------------------
#  RELEASE
# -----------------------------------------------------------------------------

release: all
	mkdir -p $(artifacts_dir)
	(cd $(linux_dir) && zip staging.zip $(bin_name))
	(cd $(darwin_dir) && zip staging.zip $(bin_name))
	(cd $(windows_dir) && zip staging.zip $(bin_name).exe)
	mv $(linux_dir)/staging.zip $(artifacts_dir)/$(zip_linux)
	mv $(darwin_dir)/staging.zip $(artifacts_dir)/$(zip_darwin)
	mv $(windows_dir)/staging.zip $(artifacts_dir)/$(zip_windows)

release-dry: release

# -----------------------------------------------------------------------------
#  CLEANUP
# -----------------------------------------------------------------------------

clean:
	rm -rf $(build_dir)
