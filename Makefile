deps:
	go get github.com/golang/lint/golint
	go get github.com/mitchellh/gox

test:
	go vet $(go list ./... | grep -v vendor)
	go test $(go list ./... | grep -v '/vendor/')

build: 
	gox -output "chinchilla" -osarch="linux/amd64"

docker: build
	docker build -t benschw/chinchilla .

package:
	gox -ldflags "-X main.Version=$TRAVIS_BUILD_NUMBER" -output "chinchilla_{{.OS}}_{{.Arch}}" -osarch="linux/amd64"
	gzip chinchilla_linux_amd64
	mkdir -p dist release
	cp chinchilla_linux_amd64.gz dist/chinchilla_linux_amd64_latest.gz
	cp chinchilla_linux_amd64.gz release/chinchilla_linux_amd64_$(git describe --tags).gz

publish: docker
	docker push benschw/chinchilla


