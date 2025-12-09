########################################
# Stage 1: Build Python Dependencies
########################################
FROM python:3.12-slim AS builder

WORKDIR /app

# Install build tools (for building wheels if needed)
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
    && rm -rf /var/lib/apt/lists/*

# Upgrade pip
RUN pip install --upgrade pip

# Install dependencies into /wheels directory
COPY requirements.txt .
RUN pip wheel --no-cache-dir --wheel-dir /wheels -r requirements.txt


########################################
# Stage 2: Runtime Image
########################################
FROM python:3.12-slim

WORKDIR /app

# ---- Create non-root user & group ----
RUN groupadd -g 101 appgroup && \
    useradd -u 100 -g 101 -m -s /usr/sbin/nologin appuser

# Install only wheel files from builder stage (much smaller)
COPY --from=builder /wheels /wheels
RUN pip install --no-cache-dir /wheels/*

# Copy application code
COPY . .

# Fix file permissions for non-root user
RUN chown -R appuser:appgroup /app

# Switch to secure user
USER appuser

EXPOSE 5000

CMD ["python", "-m", "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "5000"]