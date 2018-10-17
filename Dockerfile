# This is an example operator.
#
# The standard name for its image is enj/example-operator
#
FROM openshift/origin-release:golang-1.10
COPY . /go/src/github.com/enj/example-operator
RUN cd /go/src/github.com/enj/example-operator && go build ./cmd/example

FROM centos:7
COPY --from=0 /go/src/github.com/enj/example-operator/example /usr/bin/example
