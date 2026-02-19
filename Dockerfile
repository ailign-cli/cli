FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY ailign /usr/local/bin/ailign
ENTRYPOINT ["ailign"]
