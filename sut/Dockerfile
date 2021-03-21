FROM alpine

RUN apk add --no-cache openssh sshpass nmap
ADD test.sh /test.sh
RUN chmod +x /test.sh
CMD ["/test.sh"]