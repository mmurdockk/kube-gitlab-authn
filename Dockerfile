FROM golang:1.10-alpine as builder

ARG repo=github.comu/xuwang/kube-gitlab-authn
RUN apk --update add git
ADD . ${GOPATH}/src/${repo}
RUN go get github.com/xanzy/go-gitlab && cd ${GOPATH}/src/github.com/xanzy/go-gitlab && git checkout v0.10.5
RUN cd ${GOPATH}/src/${repo} && go get ./...

FROM alpine:3.5
RUN apk --no-cache --update add ca-certificates
COPY --from=builder /go/bin/kube-gitlab-authn /kube-gitlab-authn
CMD ["/kube-gitlab-authn"]
