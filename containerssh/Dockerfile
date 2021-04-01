ARG CONTAINERSSH_VERSION
ARG GITHUB_TOKEN
ARG SOURCE_REPO=https://github.com/ContainerSSH/ContainerSSH

FROM alpine AS download
ARG CONTAINERSSH_VERSION
ARG GITHUB_TOKEN
ARG SOURCE_REPO
RUN if [ -z "${CONTAINERSSH_VERSION}" ]; then echo "Error: No CONTAINERSSH_VERSION specified." >&2; exit 1; fi
RUN if [ -z "${GITHUB_TOKEN}" ]; then echo "Warning: No GITHUB_TOKEN specified, build may fail." >&2; fi
RUN apk add --no-cache curl
RUN mkdir -p /containerssh && \
    cd /containerssh && \
    adduser --uid 1022 --disabled-password --system --no-create-home containerssh && \
    chown containerssh /containerssh
# Drop privileges for download
USER 1022:1022
RUN cd /containerssh && \
    URL=${SOURCE_REPO}/releases/download/${CONTAINERSSH_VERSION}/containerssh_${CONTAINERSSH_VERSION/v/}_linux_amd64.tar.gz && \
    if [ -n "${CONTAINERSSH_VERSION}" ]; then \
        curl -L -o containerssh.tar.gz --header 'authorization: Bearer ${GITHUB_TOKEN}' ${URL}; \
    else \
        curl -L -o containerssh.tar.gz ${URL}; \
    fi && \
    tar -xvvzf containerssh.tar.gz && \
    rm containerssh.tar.gz
USER 0:0
RUN chown -R root:root /containerssh

FROM alpine
RUN apk upgrade --no-cache && apk add --no-cache 'libssl1.1>=1.1.1k-r0' 'libcrypto1.1>=1.1.1k-r0'
COPY --from=download /containerssh/containerssh /
COPY --from=download /containerssh/LICENSE.md /
COPY --from=download /containerssh/NOTICE.md /
RUN chmod +x /containerssh
ENTRYPOINT ["/containerssh"]
CMD ["--config", "/etc/containerssh/config.yaml"]
VOLUME /etc/containerssh
VOLUME /var/secrets
# Run as user 1022 to avoid running as root
USER 1022:1022
# Expose SSH port
EXPOSE 2222
# Expose metrics port (disabled by default)
EXPOSE 9100
