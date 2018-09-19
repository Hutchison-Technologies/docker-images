FROM alpine:3.8

RUN apk --update --no-cache add bash && \
    rm -rf /var/cache/*/* && \
    sed -i -e "s/bin\/ash/bin\/bash/" /etc/passwd 