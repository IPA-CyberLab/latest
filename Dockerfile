FROM gcr.io/distroless/base
COPY latest /
ENTRYPOINT ["/latest"]
