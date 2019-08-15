FROM alpine:3.10

RUN apk -U add ca-certificates

ADD main.bin.tmp /bin/main

CMD ["main"]