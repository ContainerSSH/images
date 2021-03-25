FROM alpine

RUN apk upgrade --no-cache && apk add --no-cache 'libssl1.1>=1.1.1k-r0' 'libcrypto1.1>=1.1.1k-r0'
RUN apk add --no-cache openssh sshpass nmap
ADD test.sh /test.sh
RUN chmod +x /test.sh
CMD ["/test.sh"]