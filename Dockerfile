FROM scratch
COPY kubernetes-config-collector /
ENTRYPOINT ["/kubernetes-config-collector"]
