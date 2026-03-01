# =============================================================================
# YouTube Python SDK — Multi-stage Dockerfile
# =============================================================================
#
# Stages:
#   base  — shared Python + system dependencies
#   dev   — development environment with all extras (tests, linting)
#   prod  — minimal runtime image for library consumers
#
# Build examples:
#   docker build --target dev  -t youtube-sdk:dev  .
#   docker build --target prod -t youtube-sdk:prod .

ARG PYTHON_VERSION=3.12
ARG DEBIAN_RELEASE=slim-bookworm

# -----------------------------------------------------------------------------
# Stage: base
# -----------------------------------------------------------------------------
FROM python:${PYTHON_VERSION}-${DEBIAN_RELEASE} AS base

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

WORKDIR /app

# Copy only the packaging metadata first to leverage Docker layer caching.
COPY pyproject.toml README.md LICENSE ./
COPY youtube/ ./youtube/

# -----------------------------------------------------------------------------
# Stage: dev
# -----------------------------------------------------------------------------
FROM base AS dev

# Install the package with development extras.
RUN pip install -e ".[dev]"

COPY tests/ ./tests/

# Default command runs the full test suite.
CMD ["pytest", "--tb=short", "-v"]

# -----------------------------------------------------------------------------
# Stage: prod
# -----------------------------------------------------------------------------
FROM base AS prod

# Install only the runtime dependencies.
RUN pip install .

# The SDK is a library — nothing to run by default.
# Override CMD in your own Dockerfile to use it.
CMD ["python", "-c", "import youtube; print(f'YouTube Python SDK v{youtube.__version__} ready.')"]
