# syntax = edrevo/dockerfile-plus@sha256:d234bd015db8acef1e628e012ea8815f6bf5ece61c7bf87d741c466919dd4e66
INCLUDE+ Dockerfile

# Dockerfile just for the action. The container is built with a non-root
# user and runs in root, both of which causes issues with gh workflows.
WORKDIR /github/workspace
USER 1001:1001
