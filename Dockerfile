FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
    && addgroup -S ailign && adduser -S -G ailign ailign
COPY ailign /usr/local/bin/ailign
USER ailign:ailign
ENTRYPOINT ["ailign"]
