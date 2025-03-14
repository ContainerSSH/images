ARG CONTAINERSSH_VERSION
ARG GITHUB_TOKEN
ARG SOURCE_REPO=https://github.com/ContainerSSH/ContainerSSH

FROM alpine AS download
ARG CONTAINERSSH_VERSION
ARG GITHUB_TOKEN
ARG SOURCE_REPO
ARG TARGETOS
ARG TARGETARCH
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
    if [ "${CONTAINERSSH_VERSION}" = "0.3.0" -o "${CONTAINERSSH_VERSION}" = "0.3.1" ]; then \
        URL=${SOURCE_REPO}/releases/download/${CONTAINERSSH_VERSION}/containerssh-authconfig_${CONTAINERSSH_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz; \
    else \
        URL=${SOURCE_REPO}/releases/download/${CONTAINERSSH_VERSION}/containerssh_${CONTAINERSSH_VERSION/v/}_${TARGETOS}_${TARGETARCH}.tar.gz; \
    fi && \
    if [ -n "${CONTAINERSSH_VERSION}" ]; then \
        curl -L -o containerssh-authconfig.tar.gz --header 'authorization: Bearer ${GITHUB_TOKEN}' ${URL}; \
    else \
        curl -L -o containerssh-authconfig.tar.gz ${URL}; \
    fi && \
    tar -xvvzf containerssh-authconfig.tar.gz && \
    rm containerssh-authconfig.tar.gz
USER 0:0
RUN chown -R root:root /containerssh

FROM alpine
RUN apk upgrade --no-cache && apk add --no-cache 'libssl3' 'libcrypto3'
COPY --from=download /containerssh/containerssh-testauthconfigserver /
COPY --from=download /containerssh/LICENSE* /
COPY --from=download /containerssh/NOTICE* /
RUN chmod +x /containerssh-testauthconfigserver
ENTRYPOINT ["/containerssh-testauthconfigserver"]
CMD []
USER 1080:1080
EXPOSE 8080
