# Multi-stage Dockerfile for Protei_Bulk C++
# Stage 1: Build environment
FROM ubuntu:22.04 AS builder

# Avoid prompts from apt
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    pkg-config \
    libboost-all-dev \
    libpqxx-dev \
    libhiredis-dev \
    libssl-dev \
    zlib1g-dev \
    postgresql-client \
    libpq-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install redis-plus-plus
WORKDIR /tmp
RUN git clone https://github.com/sewenew/redis-plus-plus.git && \
    cd redis-plus-plus && \
    mkdir build && cd build && \
    cmake .. && \
    make -j$(nproc) && \
    make install && \
    ldconfig && \
    cd /tmp && rm -rf redis-plus-plus

# Copy source code
WORKDIR /build
COPY . .

# Build the application
RUN mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc) && \
    make install

# Stage 2: Runtime environment
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Install runtime dependencies only
RUN apt-get update && apt-get install -y \
    libboost-system1.74.0 \
    libboost-thread1.74.0 \
    libboost-program-options1.74.0 \
    libpqxx-6.4 \
    libhiredis0.14 \
    libssl3 \
    zlib1g \
    postgresql-client \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy redis-plus-plus library from builder
COPY --from=builder /usr/local/lib/libredis++.so* /usr/local/lib/
RUN ldconfig

# Create application user
RUN useradd -r -u 1000 -g root -s /bin/bash protei && \
    mkdir -p /opt/protei_bulk/{config,logs,cdr,tmp} && \
    chown -R protei:root /opt/protei_bulk

# Copy binary from builder
COPY --from=builder /usr/local/bin/protei_bulk /usr/local/bin/
COPY --from=builder /build/config /opt/protei_bulk/config

WORKDIR /opt/protei_bulk

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

USER protei

EXPOSE 8080 2775

CMD ["protei_bulk", "/opt/protei_bulk/config/app.conf"]
