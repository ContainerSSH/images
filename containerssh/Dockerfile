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

# Detect architecture and download the matching binary
RUN uname -m > /tmp/arch

FROM alpine AS download-amd64
ARG CONTAINERSSH_VERSION
ARG GITHUB_TOKEN
ARG SOURCE_REPO
RUN apk add --no-cache curl
RUN mkdir -p /containerssh && adduser --uid 1022 --disabled-password --system --no-create-home containerssh
USER 1022:1022
RUN cd /containerssh && \
    URL=${SOURCE_REPO}/releases/download/${CONTAINERSSH_VERSION}/containerssh_${CONTAINERSSH_VERSION/v/}_linux_amd64.tar.gz && \
    if [ -n "${GITHUB_TOKEN}" ]; then \
        curl -L -o containerssh.tar.gz --header "authorization: Bearer ${GITHUB_TOKEN}" ${URL}; \
    else \
        curl -L -o containerssh.tar.gz ${URL}; \
    fi && \
    tar xzf containerssh.tar.gz && rm containerssh.tar.gz
USER 0:0
RUN mv /containerssh/containerssh /containerssh/containerssh-amd64

FROM alpine AS download-arm64
ARG CONTAINERSSH_VERSION
ARG GITHUB_TOKEN
ARG SOURCE_REPO
RUN apk add --no-cache curl
RUN mkdir -p /containerssh && adduser --uid 1022 --disabled-password --system --no-create-home containerssh
USER 1022:1022
RUN cd /containerssh && \
    URL=${SOURCE_REPO}/releases/download/${CONTAINERSSH_VERSION}/containerssh_${CONTAINERSSH_VERSION/v/}_linux_arm64.tar.gz && \
    if [ -n "${GITHUB_TOKEN}" ]; then \
        curl -L -o containerssh.tar.gz --header "authorization: Bearer ${GITHUB_TOKEN}" ${URL}; \
    else \
        curl -L -o containerssh.tar.gz ${URL}; \
    fi && \
    tar xzf containerssh.tar.gz && rm containerssh.tar.gz
USER 0:0
RUN mv /containerssh/containerssh /containerssh/containerssh-arm64

FROM alpine AS merge
COPY --from=download-amd64 /containerssh/containerssh-amd64 /containerssh/
COPY --from=download-arm64 /containerssh/containerssh-arm64 /containerssh/
COPY --from=download-amd64 /containerssh/LICENSE* /containerssh/
COPY --from=download-amd64 /containerssh/NOTICE* /containerssh/

FROM alpine
ARG TARGETARCH
RUN apk upgrade --no-cache && apk add --no-cache 'libssl3' 'libcrypto3'
# Copy only the binary for the target architecture
COPY --from=merge /containerssh/containerssh-${TARGETARCH} /containerssh/containerssh
COPY --from=merge /containerssh/LICENSE* /
COPY --from=merge /containerssh/NOTICE*
RUN chmod +x /containerssh && rm -rf /containerssh/containerssh-*
ENTRYPOINT ["/containerssh"]
CMD ["--config", "/etc/containerssh/config.yaml"]
VOLUME /etc/containerssh
VOLUME /var/secrets
USER 1022:1022
EXPOSE 2222
EXPOSE 9100
