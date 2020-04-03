FROM bitnami/kubectl:1.18
USER 0
RUN install_packages jq
RUN mkdir /.kube
RUN chown 1001 .kube
USER 1001
COPY ./test_internal.sh /
COPY ./testspec.yaml /
COPY --chown=1001 ./.kube/config /.kube/config
ENTRYPOINT ["/test_internal.sh"]
