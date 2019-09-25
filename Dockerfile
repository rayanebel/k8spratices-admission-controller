FROM golang:1.13.0-alpine as build

WORKDIR /app

ADD . /app

RUN CGO_ENABLED=0 GOOS=linux go build -o k8spractices-admission-controller

FROM golang:1.13.0-alpine

USER 1

WORKDIR /app

EXPOSE 8443

COPY --from=build /app/k8spractices-admission-controller /app

ENTRYPOINT ["/app/k8spractices-admission-controller"]