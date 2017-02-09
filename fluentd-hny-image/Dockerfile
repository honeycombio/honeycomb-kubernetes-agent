FROM gcr.io/google_containers/ubuntu-slim:0.6

# Ensure there are enough file descriptors for running Fluentd.
RUN ulimit -n 65536

# Disable prompts from apt.
ENV DEBIAN_FRONTEND noninteractive

# Copy the Fluentd configuration file.
COPY td-agent.conf /etc/td-agent/td-agent.conf

RUN apt-get update && apt-get install -y curl

RUN curl https://packages.treasuredata.com/GPG-KEY-td-agent | apt-key add -
RUN echo "deb http://packages.treasuredata.com/2/ubuntu/xenial/ xenial contrib" > /etc/apt/sources.list.d/treasure-data.list
RUN apt-get update && apt-get install -y \
    ca-certificates \
    make \
    g++ \
    sudo \
    bash \
    td-agent

# Change the default user and group to root.
# Needed to allow access to /var/log/docker/... files.
RUN sed -i -e "s/USER=td-agent/USER=root/" -e "s/GROUP=td-agent/GROUP=root/" /etc/init.d/td-agent

RUN td-agent-gem install --no-document fluent-plugin-kubernetes_metadata_filter -v 0.24.0
RUN td-agent-gem install --no-document fluent-plugin-honeycomb -v 0.2.0


# Remove docs and postgres references
RUN rm -rf /opt/td-agent/embedded/share/doc \
  /opt/td-agent/embedded/share/gtk-doc \
  /opt/td-agent/embedded/lib/postgresql \
  /opt/td-agent/embedded/bin/postgres \
  /opt/td-agent/embedded/share/postgresql

RUN apt-get remove -y make g++ && apt-get autoremove -y && apt-get clean -y

RUN rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV LD_PRELOAD /opt/td-agent/embedded/lib/libjemalloc.so

# Run the Fluentd service.
ENTRYPOINT ["td-agent"]
