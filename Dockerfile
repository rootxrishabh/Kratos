FROM ubuntu:latest

ENV DIGITAL_OCEAN_TOKEN=dop_v1_b8673b9a23b0bfb63ecba03ce2b03db9908006ed592f24128c98b8d669184762

COPY kratos /usr/local/bin/kratos

ENTRYPOINT [ "kratos" ]