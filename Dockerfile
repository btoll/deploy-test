FROM golang:1.21.1-bookworm AS build

WORKDIR /app
COPY . /app/

RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build

#FROM gcr.io/distroless/base-debian11 AS final
FROM golang:1.21.1-bookworm AS final

RUN groupadd --gid 1000 noroot \
    && useradd \
    --create-home \
    --home-dir /home/noroot \
    --uid 1000 \
    --gid 1000 \
    # https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user
    --no-log-init noroot

WORKDIR /home/noroot

RUN mkdir /home/noroot/.ssh && \
    { \ echo "Host *"; \
        echo "    IdentityFile=/home/noroot/.ssh/foo_test"; \
        echo "    StrictHostKeyChecking=no"; \
    } >> /home/noroot/.ssh/config && \
    touch /home/noroot/.ssh/known_hosts && \
    chown -R noroot: /home/noroot

COPY foo_test known_hosts /home/noroot/.ssh/
COPY --from=build /app/deploy-test .

RUN chown -R noroot:noroot /home/noroot

USER noroot

ENTRYPOINT ["/home/noroot/deploy-test"]

