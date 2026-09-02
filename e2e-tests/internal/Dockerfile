FROM alpine/k8s:1.21.13
USER 0
RUN apk add --no-cache netcat-openbsd
RUN mkdir /.kube
RUN chown 1001 /.kube
USER 1001
COPY ./test_internal.sh /
COPY ./testspec.yaml /
COPY --chown=1001 ./.kube/config /.kube/config
ENTRYPOINT ["/test_internal.sh"]
