FROM alpine

RUN apk upgrade --no-cache && apk add --no-cache 'libssl3' 'libcrypto3'
RUN apk add --no-cache openssh sshpass nmap
ADD test.sh /test.sh
RUN chmod +x /test.sh
CMD ["/test.sh"]
