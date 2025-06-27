# GoReleaser Dockerfile
# This Dockerfile is designed to work with GoReleaser's build context
# which includes only the pre-built binary and extra files

FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy pre-built binary from GoReleaser context
COPY t212-taxes .

# Copy configuration files from GoReleaser context
COPY config.yaml .

# Create data directory and set permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app && \
    chmod +x /app/t212-taxes

# Switch to non-root user
USER appuser

# Expose port (if needed for future web interface)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ./t212-taxes --help || exit 1

# Default command
ENTRYPOINT ["./t212-taxes"]
CMD ["--help"] 