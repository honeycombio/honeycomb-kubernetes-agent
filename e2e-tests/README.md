This directory contains scripts to run a simple end-to-end smoke test of the
agent inside Minikube. It starts:

- a sample nginx service
- the Honeycomb agent, configured to slurp logs from the nginx service
- a mock Honeycomb API server that the agent will talk to.

It then issues a request to nginx and checks that the mock API server actually
gets a corresponding event.
