FROM alpine:latest

RUN apk add  --update ca-certificates

RUN mkdir -p /etc/simplcheck

COPY ./simplcheck /simplcheck
RUN chmod u+x /simplcheck

CMD ["/simplcheck", "-conf", "/etc/simplcheck/default.json"]
