# syntax = edrevo/dockerfile-plus
INCLUDE+ Dockerfile

# Dockerfile just for the action. The container is built with a non-root
# user and runs in root, both of which causes issues with gh workflows.
WORKDIR /github/workspace
USER 1001:1001
