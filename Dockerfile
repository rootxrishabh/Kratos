FROM ubuntu:latest

ENV DIGITAL_OCEAN_TOKEN=CODESPACE_SECRET

COPY kratos /usr/local/bin/kratos

ENTRYPOINT [ "kratos" ]
